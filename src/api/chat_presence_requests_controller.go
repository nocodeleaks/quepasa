package api

import (
	"context"
	"strings"
	"time"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type ChatPresenceRequestExtended struct {
	ChatPresenceRequest

	ctx    context.Context
	Cancel context.CancelFunc
}

func Exec(ctx context.Context, request *ChatPresenceRequest, server *models.QpWhatsappServer) {
	logentry := server.GetLogger()
	logentry = logentry.WithField(LogFields.ChatId, request.ChatId)

	logentry.Tracef("background chat presence update, duration: %d ms\n", request.Duration)
	defer logentry.Trace("background chat presence update finished")

	// Calculate total duration and end time
	duration := time.Duration(request.Duration) * time.Millisecond
	endTime := time.Now().UTC().Add(duration)

	// Use shorter sleep intervals to check for cancellation more frequently
	const checkInterval = 500 * time.Millisecond // 500 milliseconds for cancellation check

	logentry.Debugf("background chat presence update, with presence type: %s...", request.Type)

	for time.Now().UTC().Before(endTime) {
		select {
		case <-ctx.Done():
			logentry.Debug("background chat presence update received cancellation signal (replaced by new request)")
			return // Don't send paused - new request will handle it
		case <-time.After(checkInterval):
			// Just wait - presence indicator was already sent by the controller
			// We only refresh to check for cancellation
		}
	}

	// Only send paused indicator if we reached timeout (not cancelled)
	logentry.Trace("background chat presence timeout reached, sending paused indicator")
	err := server.SendChatPresence(request.ChatId, whatsapp.WhatsappChatPresenceTypePaused)
	if err != nil {
		logentry.Errorf("failed to send paused indicator: %v", err)
	} else {
		logentry.Debug("sent paused indicator after presence timeout")
	}
}

type ChatPresenceRequests struct {
	models.QpCache
}

var ChatPresenceRequestsController ChatPresenceRequests

//#region MESSAGES

func (source *ChatPresenceRequests) Append(request *ChatPresenceRequest, server *models.QpWhatsappServer) {
	ctx, cancel := context.WithCancel(context.Background())

	value := &ChatPresenceRequestExtended{
		ChatPresenceRequest: *request,
		ctx:                 ctx,
		Cancel:              cancel,
	}

	chatid := value.ChatId

	// ensure that is an uppercase string before save
	normalizedId := strings.ToUpper(chatid)

	expiration := time.Now().UTC().Add(time.Duration(request.Duration) * time.Millisecond)

	item := models.QpCacheItem{
		Key:        normalizedId,
		Value:      value,
		Expiration: expiration,
	}

	// set the item in the cache
	ok := source.SetCacheItem(item, "chatpresence")
	if ok {
		go Exec(ctx, request, server)
	}
}

func (source *ChatPresenceRequests) Cancel(chatid string) bool {

	// ensure that is an uppercase string before save
	normalizedId := strings.ToUpper(chatid)

	cached, found := source.GetAny(normalizedId)
	if !found {
		return false
	}

	extended, ok := cached.(*ChatPresenceRequestExtended)
	if !ok {
		return false
	}

	extended.Cancel()
	return true
}
