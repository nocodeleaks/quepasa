package cable

import (
	"context"
	"strings"
	"time"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type cablePresenceRequest struct {
	ChatID   string
	Type     whatsapp.WhatsappChatPresenceType
	Duration uint

	ctx    context.Context
	cancel context.CancelFunc
}

type cablePresenceRequestsController struct {
	models.QpCache
}

var cablePresenceRequests cablePresenceRequestsController

func (controller *cablePresenceRequestsController) Append(chatID string, presenceType whatsapp.WhatsappChatPresenceType, duration uint, server *models.QpWhatsappServer) {
	ctx, cancel := context.WithCancel(context.Background())

	request := &cablePresenceRequest{
		ChatID:   chatID,
		Type:     presenceType,
		Duration: duration,
		ctx:      ctx,
		cancel:   cancel,
	}

	expiration := time.Now().UTC().Add(time.Duration(duration) * time.Millisecond)
	item := models.QpCacheItem{
		Key:        normalizePresenceChatID(chatID),
		Value:      request,
		Expiration: expiration,
	}

	if controller.SetCacheItem(item, "cable-chatpresence") {
		go controller.exec(ctx, request, server)
	}
}

func (controller *cablePresenceRequestsController) Cancel(chatID string) bool {
	cached, found := controller.GetAny(normalizePresenceChatID(chatID))
	if !found {
		return false
	}

	request, ok := cached.(*cablePresenceRequest)
	if !ok {
		return false
	}

	request.cancel()
	return true
}

func (controller *cablePresenceRequestsController) exec(ctx context.Context, request *cablePresenceRequest, server *models.QpWhatsappServer) {
	duration := time.Duration(request.Duration) * time.Millisecond
	endTime := time.Now().UTC().Add(duration)
	const checkInterval = 500 * time.Millisecond

	for time.Now().UTC().Before(endTime) {
		select {
		case <-ctx.Done():
			return
		case <-time.After(checkInterval):
		}
	}

	_ = server.SendChatPresence(request.ChatID, whatsapp.WhatsappChatPresenceTypePaused)
}

func normalizePresenceChatID(chatID string) string {
	return strings.ToUpper(strings.TrimSpace(chatID))
}
