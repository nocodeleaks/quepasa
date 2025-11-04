package whatsmeow

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	environment "github.com/nocodeleaks/quepasa/environment"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types"
)

// WakeUpScheduler manages scheduled presence activation at a specific hour of day
// It monitors configured wake-up hour and automatically sets presence to online
// for a configured duration, then returns to offline
type WakeUpScheduler struct {
	mu              sync.RWMutex
	connection      *WhatsmeowConnection
	ctx             context.Context
	cancel          context.CancelFunc
	wakeUpTime      time.Time // single wake-up time for today
	duration        time.Duration
	enabled         bool
	onlinePresence  types.Presence
	offlinePresence types.Presence
}

// NewWakeUpScheduler creates a new wake-up scheduler
func NewWakeUpScheduler(connection *WhatsmeowConnection) *WakeUpScheduler {
	wakeUpHourStr := environment.Settings.WhatsApp.WakeUpHour
	duration := environment.Settings.WhatsApp.WakeUpDuration
	
	// Parse wake-up hour from configuration
	wakeUpTime, err := parseWakeUpHour(wakeUpHourStr)
	enabled := err == nil && !wakeUpTime.IsZero()
	
	if err != nil && len(wakeUpHourStr) > 0 {
		logentry := connection.GetLogger()
		logentry.Warnf("failed to parse WAKEUP_HOUR '%s': %v", wakeUpHourStr, err)
	}
	
	// Default duration if not set or invalid
	if duration <= 0 {
		duration = 10 // 10 seconds default
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	scheduler := &WakeUpScheduler{
		connection:      connection,
		wakeUpTime:      wakeUpTime,
		duration:        time.Duration(duration) * time.Second,
		enabled:         enabled,
		onlinePresence:  types.PresenceAvailable,
		offlinePresence: types.PresenceUnavailable,
		ctx:             ctx,
		cancel:          cancel,
	}
	
	// Start scheduler if enabled
	if scheduler.enabled {
		go scheduler.run()
	}
	
	return scheduler
}

// parseWakeUpHour parses wake-up hour string in format "0-23"
// Returns time for today at the specified hour
func parseWakeUpHour(hourStr string) (time.Time, error) {
	if hourStr == "" {
		return time.Time{}, nil
	}
	
	hourStr = strings.TrimSpace(hourStr)
	
	// Parse hour as integer (0-23)
	var hour int
	_, err := fmt.Sscanf(hourStr, "%d", &hour)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid hour format '%s': %v (expected 0-23)", hourStr, err)
	}
	
	// Validate hour range
	if hour < 0 || hour > 23 {
		return time.Time{}, fmt.Errorf("hour '%d' out of range (must be 0-23)", hour)
	}
	
	// Create time for today at the specified hour (minute=0, second=0)
	now := time.Now()
	todayTime := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	
	return todayTime, nil
}

// IsEnabled returns whether the scheduler is enabled
func (ws *WakeUpScheduler) IsEnabled() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.enabled
}

// run is the main scheduler loop that checks for wake-up hour
func (ws *WakeUpScheduler) run() {
	logentry := ws.connection.GetLogger()
	logentry.Infof("wake-up scheduler started for hour: %d:00", ws.wakeUpTime.Hour())
	
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes (once per day execution)
	defer ticker.Stop()
	
	for {
		select {
		case <-ws.ctx.Done():
			logentry.Info("wake-up scheduler stopped")
			return
		case <-ticker.C:
			ws.checkAndExecuteWakeUp()
		}
	}
}

// checkAndExecuteWakeUp checks if current time matches the wake-up hour
func (ws *WakeUpScheduler) checkAndExecuteWakeUp() {
	ws.mu.RLock()
	wakeUpTime := ws.wakeUpTime
	duration := ws.duration
	ws.mu.RUnlock()
	
	now := time.Now()
	logentry := ws.connection.GetLogger()
	
	// Check if we're within the wake-up window (within 10 minutes of scheduled hour)
	// Since we check every 5 minutes, 10-minute window ensures we don't miss it
	diff := now.Sub(wakeUpTime)
	
	// If the wake-up time is in the past but less than 10 minutes ago, execute
	if diff >= 0 && diff < 10*time.Minute {
		logentry.Infof("wake-up triggered at scheduled hour: %d:00", wakeUpTime.Hour())
		ws.executeWakeUp(duration, logentry)
		
		// Update wake-up time to tomorrow to avoid re-triggering
		ws.mu.Lock()
		ws.wakeUpTime = ws.wakeUpTime.Add(24 * time.Hour)
		logentry.Debugf("next wake-up scheduled for: %s", ws.wakeUpTime.Format("2006-01-02 15:00"))
		ws.mu.Unlock()
	}
}

// executeWakeUp activates presence and schedules deactivation
func (ws *WakeUpScheduler) executeWakeUp(duration time.Duration, logentry *log.Entry) {
	logentry.Infof("executing wake-up: setting presence to online for %v", duration)
	
	// Set presence to online
	ws.sendPresence(ws.onlinePresence, logentry, "scheduled wake-up")
	
	// Schedule return to offline
	time.AfterFunc(duration, func() {
		ws.sendPresence(ws.offlinePresence, logentry, "wake-up duration expired")
		logentry.Infof("wake-up completed: presence set to offline after %v", duration)
	})
}

// sendPresence sends presence update to WhatsApp
func (ws *WakeUpScheduler) sendPresence(presence types.Presence, logentry *log.Entry, reason string) {
	if ws.connection.Client == nil || ws.connection.Client.Store == nil {
		logentry.Warn("cannot send presence: client not available")
		return
	}
	
	if len(ws.connection.Client.Store.PushName) == 0 {
		logentry.Debug("cannot send presence: push name not set")
		return
	}
	
	SendPresence(ws.connection.Client, presence, reason, logentry)
}

// Stop stops the scheduler
func (ws *WakeUpScheduler) Stop() {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	
	if ws.cancel != nil {
		ws.cancel()
		ws.cancel = nil
	}
}

// Dispose cleans up resources
func (ws *WakeUpScheduler) Dispose() {
	ws.Stop()
}
