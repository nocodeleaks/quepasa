package whatsmeow

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	waMmsRetry "go.mau.fi/whatsmeow/proto/waMmsRetry"
	types "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

const mediaRetryEntryTTL = 15 * time.Minute

type PendingMediaRetry struct {
	Message   *whatsapp.WhatsappMessage
	MediaKey  []byte
	CreatedAt time.Time
}

func normalizeRetryID(messageID string) string {
	return strings.ToUpper(strings.TrimSpace(messageID))
}

func (source *WhatsmeowConnection) handleDownloadMediaRetry(imsg whatsapp.IWhatsappMessage, downloadable whatsmeow.DownloadableMessage, downloadErr error) ([]byte, error) {
	if !errors.Is(downloadErr, whatsmeow.ErrMediaDownloadFailedWith403) &&
		!errors.Is(downloadErr, whatsmeow.ErrMediaDownloadFailedWith404) &&
		!errors.Is(downloadErr, whatsmeow.ErrMediaDownloadFailedWith410) {
		return nil, downloadErr
	}

	logentry := source.GetLogger().WithField(LogFields.MessageId, imsg.GetId())

	messageInfo, err := source.getMessageInfoForMediaRetry(imsg)
	if err != nil {
		logentry.Warnf("media retry skipped: cannot build message info: %v", err)
		return nil, downloadErr
	}

	mediaKey := getMediaKeyFromDownloadable(downloadable)
	if len(mediaKey) == 0 {
		logentry.Warn("media retry skipped: media key not found")
		return nil, downloadErr
	}

	err = source.Client.SendMediaRetryReceipt(context.Background(), messageInfo, mediaKey)
	if err != nil {
		logentry.Warnf("media retry request failed: %v", err)
		return nil, fmt.Errorf("%w (media retry request failed: %v)", downloadErr, err)
	}

	err = source.storePendingMediaRetry(imsg, mediaKey)
	if err != nil {
		logentry.Warnf("media retry requested but pending state could not be stored: %v", err)
		return nil, fmt.Errorf("%w (media retry requested, but local pending state failed: %v)", downloadErr, err)
	}

	logentry.Infof("media retry requested after download failure: %v", downloadErr)
	return nil, fmt.Errorf("%w (media retry requested, webhook will be triggered if media becomes available)", downloadErr)
}

func (source *WhatsmeowConnection) getMessageInfoForMediaRetry(imsg whatsapp.IWhatsappMessage) (*types.MessageInfo, error) {
	waMsg, ok := imsg.(*whatsapp.WhatsappMessage)
	if !ok || waMsg == nil {
		return nil, fmt.Errorf("message type does not expose metadata for media retry")
	}

	if info, ok := waMsg.InfoForHistory.(types.MessageInfo); ok {
		copied := info
		return &copied, nil
	}

	if info, ok := waMsg.InfoForHistory.(*types.MessageInfo); ok && info != nil {
		copied := *info
		return &copied, nil
	}

	chatJID, err := types.ParseJID(waMsg.Chat.Id)
	if err != nil {
		return nil, fmt.Errorf("cannot parse chat id: %w", err)
	}

	sender := chatJID
	if waMsg.Participant != nil && len(waMsg.Participant.Id) > 0 {
		senderJID, senderErr := types.ParseJID(waMsg.Participant.Id)
		if senderErr == nil {
			sender = senderJID
		}
	}

	info := &types.MessageInfo{
		MessageSource: types.MessageSource{
			Chat:     chatJID,
			Sender:   sender,
			IsFromMe: waMsg.FromMe,
			IsGroup:  waMsg.FromGroup(),
		},
		ID:        types.MessageID(waMsg.Id),
		Timestamp: waMsg.Timestamp,
	}
	return info, nil
}

func getMediaKeyFromDownloadable(downloadable whatsmeow.DownloadableMessage) []byte {
	switch media := downloadable.(type) {
	case *waE2E.ImageMessage:
		return media.GetMediaKey()
	case *waE2E.VideoMessage:
		return media.GetMediaKey()
	case *waE2E.AudioMessage:
		return media.GetMediaKey()
	case *waE2E.DocumentMessage:
		return media.GetMediaKey()
	case *waE2E.StickerMessage:
		return media.GetMediaKey()
	default:
		return nil
	}
}

func (source *WhatsmeowConnection) storePendingMediaRetry(imsg whatsapp.IWhatsappMessage, mediaKey []byte) error {
	waMsg, ok := imsg.(*whatsapp.WhatsappMessage)
	if !ok || waMsg == nil {
		return fmt.Errorf("message type does not support pending retry storage")
	}

	if len(waMsg.Id) == 0 {
		return fmt.Errorf("empty message id")
	}

	source.cleanupPendingMediaRetries()

	copiedKey := make([]byte, len(mediaKey))
	copy(copiedKey, mediaKey)

	entry := &PendingMediaRetry{
		Message:   waMsg,
		MediaKey:  copiedKey,
		CreatedAt: time.Now().UTC(),
	}

	source.mediaRetryPending.Store(normalizeRetryID(waMsg.Id), entry)
	return nil
}

func (source *WhatsmeowConnection) loadAndDeletePendingMediaRetry(messageID string) (*PendingMediaRetry, bool) {
	normalizedID := normalizeRetryID(messageID)
	value, loaded := source.mediaRetryPending.LoadAndDelete(normalizedID)
	if !loaded {
		return nil, false
	}

	entry, ok := value.(*PendingMediaRetry)
	if !ok || entry == nil {
		return nil, false
	}

	if time.Since(entry.CreatedAt) > mediaRetryEntryTTL {
		return nil, false
	}

	return entry, true
}

func (source *WhatsmeowConnection) cleanupPendingMediaRetries() {
	now := time.Now().UTC()
	source.mediaRetryPending.Range(func(key, value interface{}) bool {
		entry, ok := value.(*PendingMediaRetry)
		if !ok || entry == nil || now.Sub(entry.CreatedAt) > mediaRetryEntryTTL {
			source.mediaRetryPending.Delete(key)
		}
		return true
	})
}

func (handler *WhatsmeowHandlers) OnMediaRetry(evt events.MediaRetry) {
	logentry := handler.GetLogger().WithField(LogFields.MessageId, evt.MessageID)
	conn := handler.WhatsmeowConnection
	if conn == nil {
		logentry.Warn("media retry ignored: nil connection")
		return
	}

	pending, found := conn.loadAndDeletePendingMediaRetry(string(evt.MessageID))
	if !found {
		logentry.Debug("media retry received with no pending context")
		return
	}

	retryData, err := whatsmeow.DecryptMediaRetryNotification(&evt, pending.MediaKey)
	if err != nil {
		if errors.Is(err, whatsmeow.ErrMediaNotAvailableOnPhone) {
			logentry.Warn("media retry failed: media is not available on phone")
			return
		}
		logentry.Warnf("media retry decrypt failed: %v", err)
		return
	}

	if retryData.GetResult() != waMmsRetry.MediaRetryNotification_SUCCESS {
		logentry.Warnf("media retry finished with non-success result: %v", retryData.GetResult())
		return
	}

	if pending.Message == nil {
		logentry.Warn("media retry pending message is nil")
		return
	}

	waMessage, ok := pending.Message.Content.(*waE2E.Message)
	if !ok || waMessage == nil {
		logentry.Warnf("media retry pending content has unsupported type: %T", pending.Message.Content)
		return
	}

	applyDirectPathToMediaMessage(waMessage, retryData.GetDirectPath())
	downloadable := GetDownloadableMessage(waMessage)
	if downloadable == nil {
		logentry.Warn("media retry failed: message is not downloadable after direct path update")
		return
	}

	data, err := conn.Client.Download(context.Background(), downloadable)
	if err != nil {
		logentry.Warnf("media retry download still failing: %v", err)
		return
	}

	retryMessage := cloneMessageForMediaRetryDispatch(handler, pending.Message, data, evt.MessageID)
	if retryMessage == nil {
		logentry.Warn("media retry dispatch skipped: failed to build retry message")
		return
	}

	if handler.WAHandlers == nil || handler.WAHandlers.IsInterfaceNil() {
		logentry.Warn("media retry dispatch skipped: internal handler not attached")
		return
	}

	logentry.Infof("media retry succeeded, dispatching as new message id=%s", retryMessage.Id)
	handler.WAHandlers.Message(retryMessage, "mediaretry")
}

func applyDirectPathToMediaMessage(message *waE2E.Message, directPath string) {
	if message == nil || len(directPath) == 0 {
		return
	}

	directPathPtr := proto.String(directPath)

	switch {
	case message.ImageMessage != nil:
		message.ImageMessage.DirectPath = directPathPtr
	case message.VideoMessage != nil:
		message.VideoMessage.DirectPath = directPathPtr
	case message.AudioMessage != nil:
		message.AudioMessage.DirectPath = directPathPtr
	case message.DocumentMessage != nil:
		message.DocumentMessage.DirectPath = directPathPtr
	case message.StickerMessage != nil:
		message.StickerMessage.DirectPath = directPathPtr
	case message.PtvMessage != nil:
		message.PtvMessage.DirectPath = directPathPtr
	}
}

func cloneMessageForMediaRetryDispatch(handler *WhatsmeowHandlers, source *whatsapp.WhatsappMessage, data []byte, retryMessageID types.MessageID) *whatsapp.WhatsappMessage {
	if source == nil {
		return nil
	}

	cloned := *source
	originalID := source.Id
	newID := strings.ToUpper(fmt.Sprintf("%s-MEDIARETRY-%s", originalID, uuid.NewString()))
	if handler != nil && handler.Client != nil {
		generatedID := strings.ToUpper(handler.Client.GenerateMessageID())
		if len(generatedID) > 0 {
			newID = fmt.Sprintf("%s-MEDIARETRY-%s", strings.ToUpper(originalID), generatedID)
		}
	} else {
		newID = strings.ToUpper(fmt.Sprintf("%s-MEDIARETRY-%s", originalID, uuid.NewString()))
	}

	cloned.Id = newID
	cloned.FromHistory = false
	cloned.Edited = true
	cloned.Timestamp = time.Now().UTC()
	cloned.Info = map[string]interface{}{
		"event":               "media_retry_downloaded",
		"original_message_id": originalID,
		"media_retry_id":      string(retryMessageID),
		"timestamp":           time.Now().UTC().Format(time.RFC3339),
	}

	if source.Attachment != nil {
		attachment := *source.Attachment
		cloned.Attachment = &attachment
		cloned.Attachment.SetContent(&data)
		if cloned.Attachment.FileLength == 0 {
			cloned.Attachment.FileLength = uint64(len(data))
		}
	}

	return &cloned
}
