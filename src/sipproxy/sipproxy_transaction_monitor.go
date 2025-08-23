package sipproxy

import (
	"sync"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	log "github.com/sirupsen/logrus"
)

// SIPCallAcceptedCallback é chamado quando uma chamada SIP é aceita (200 OK)
type SIPCallAcceptedCallback func(callID, fromPhone, toPhone string, response *sip.Response)

// SIPCallRejectedCallback é chamado quando uma chamada SIP é rejeitada (>=400)
type SIPCallRejectedCallback func(callID, fromPhone, toPhone string, response *sip.Response)

// SIPTransactionMonitor handles monitoring of SIP transactions and responses
type SIPTransactionMonitor struct {
	logger              *log.Entry
	client              *sipgo.Client
	activeSessions      map[string]*TransactionSession
	mutex               sync.RWMutex
	callAcceptedHandler SIPCallAcceptedCallback
	callRejectedHandler SIPCallRejectedCallback
}

// TransactionSession represents an active SIP transaction
type TransactionSession struct {
	CallID        string
	FromPhone     string
	ToPhone       string
	TransactionID string
	StartTime     time.Time
	ResponseCount int
	LastResponse  *sip.Response
}

// NewSIPTransactionMonitor creates a new SIP transaction monitor
func NewSIPTransactionMonitor(logger *log.Entry) *SIPTransactionMonitor {
	return &SIPTransactionMonitor{
		logger:         logger,
		activeSessions: make(map[string]*TransactionSession),
	}
}

// SetCallbacks sets the callback handlers for call events
func (stm *SIPTransactionMonitor) SetCallbacks(acceptedHandler SIPCallAcceptedCallback, rejectedHandler SIPCallRejectedCallback) {
	stm.callAcceptedHandler = acceptedHandler
	stm.callRejectedHandler = rejectedHandler
}

// SetClient sets the SIP client for transaction monitoring
func (stm *SIPTransactionMonitor) SetClient(client *sipgo.Client) {
	stm.client = client
}

// MonitorTransaction starts monitoring a SIP transaction
func (stm *SIPTransactionMonitor) MonitorTransaction(callID, fromPhone, toPhone, transactionID string) {
	stm.mutex.Lock()
	session := &TransactionSession{
		CallID:        callID,
		FromPhone:     fromPhone,
		ToPhone:       toPhone,
		TransactionID: transactionID,
		StartTime:     time.Now(),
		ResponseCount: 0,
	}
	stm.activeSessions[callID] = session
	stm.mutex.Unlock()

	stm.logger.Infof("🔍 Monitoring SIP transaction for call %s", callID)
	stm.logger.Infof("🔍🔍🔍 DEBUG: Starting transaction monitoring goroutine for call %s", callID)
	stm.logger.Infof("🔍 Transaction state: %s", transactionID)

	// Start monitoring in a separate goroutine
	go stm.monitorTransactionResponses(callID)
}

// monitorTransactionResponses monitors responses for a specific transaction
func (stm *SIPTransactionMonitor) monitorTransactionResponses(callID string) {
	timeout := 30 * time.Second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(startTime)
			if elapsed >= timeout {
				stm.mutex.RLock()
				session, exists := stm.activeSessions[callID]
				stm.mutex.RUnlock()

				if exists {
					stm.logger.Warnf("⏰ Transaction timeout for call %s after %d seconds (received %d responses)",
						callID, int(timeout.Seconds()), session.ResponseCount)

					stm.mutex.Lock()
					delete(stm.activeSessions, callID)
					stm.mutex.Unlock()
				}
				return
			}

			// Check for responses (this would be enhanced with actual SIP response monitoring)
			stm.checkForResponses(callID)
		}
	}
}

// checkForResponses checks for SIP responses (placeholder for actual implementation)
func (stm *SIPTransactionMonitor) checkForResponses(callID string) {
	// This is where we would check for actual SIP responses
	// For now, this is a placeholder that can be enhanced with actual SIP response handling
	stm.mutex.RLock()
	session, exists := stm.activeSessions[callID]
	stm.mutex.RUnlock()

	if !exists {
		return
	}

	// Log periodic status
	elapsed := time.Since(session.StartTime)
	if int(elapsed.Seconds())%10 == 0 { // Log every 10 seconds
		stm.logger.Infof("🔍 Transaction %s still active after %d seconds (responses: %d)",
			callID, int(elapsed.Seconds()), session.ResponseCount)
	}
}

// HandleResponse processes a SIP response
func (stm *SIPTransactionMonitor) HandleResponse(response *sip.Response, callID, fromPhone, toPhone string) {
	stm.mutex.Lock()
	session, exists := stm.activeSessions[callID]
	if exists {
		session.ResponseCount++
		session.LastResponse = response
	}
	stm.mutex.Unlock()

	statusCode := int(response.StatusCode)
	stm.logger.Infof("📞 SIP Response received for call %s: %d %s", callID, statusCode, response.Reason)

	// Handle different response types
	if statusCode >= 200 && statusCode < 300 {
		// Success responses (call accepted)
		stm.logger.Infof("✅ Call %s ACCEPTED with status %d", callID, statusCode)
		if stm.callAcceptedHandler != nil {
			stm.callAcceptedHandler(callID, fromPhone, toPhone, response)
		}
	} else if statusCode >= 400 {
		// Error responses (call rejected)
		stm.logger.Infof("❌ Call %s REJECTED with status %d", callID, statusCode)
		if stm.callRejectedHandler != nil {
			stm.callRejectedHandler(callID, fromPhone, toPhone, response)
		}
	} else if statusCode >= 100 && statusCode < 200 {
		// Provisional responses (call progress)
		stm.logger.Infof("📞 Call %s PROGRESS with status %d", callID, statusCode)
	}

	// Remove from active sessions if final response
	if statusCode >= 200 {
		stm.mutex.Lock()
		delete(stm.activeSessions, callID)
		stm.mutex.Unlock()
		stm.logger.Infof("🔍 Transaction monitoring completed for call %s", callID)
	}
}

// RemoveActiveCall removes a call from active monitoring
func (stm *SIPTransactionMonitor) RemoveActiveCall(callID string) {
	stm.mutex.Lock()
	delete(stm.activeSessions, callID)
	stm.mutex.Unlock()
	stm.logger.Infof("📞 Call %s removed from active monitoring", callID)
}

// GetActiveSessionsCount returns the number of active sessions
func (stm *SIPTransactionMonitor) GetActiveSessionsCount() int {
	stm.mutex.RLock()
	defer stm.mutex.RUnlock()
	return len(stm.activeSessions)
}

// Stop stops all transaction monitoring
func (stm *SIPTransactionMonitor) Stop() {
	stm.mutex.Lock()
	defer stm.mutex.Unlock()

	stm.logger.Infof("🛑 Stopping SIP Transaction Monitor...")
	stm.activeSessions = make(map[string]*TransactionSession)
	stm.logger.Infof("✅ SIP Transaction Monitor stopped")
}
