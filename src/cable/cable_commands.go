package cable

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"
)

type subscriptionCommandData struct {
	Token  string   `json:"token,omitempty"`
	Tokens []string `json:"tokens,omitempty"`
	Topic  string   `json:"topic,omitempty"`
	Topics []string `json:"topics,omitempty"`
}

type serverTokenCommandData struct {
	Token string `json:"token"`
}

type sendCommandData struct {
	Token          string                     `json:"token"`
	ID             string                     `json:"id,omitempty"`
	ChatID         string                     `json:"chatId,omitempty"`
	ChatId         string                     `json:"chatid,omitempty"`
	TrackID        string                     `json:"trackId,omitempty"`
	TrackId        string                     `json:"trackid,omitempty"`
	Text           string                     `json:"text,omitempty"`
	InReply        string                     `json:"inReply,omitempty"`
	Inreply        string                     `json:"inreply,omitempty"`
	FileName       string                     `json:"fileName,omitempty"`
	Filename       string                     `json:"filename,omitempty"`
	FileLength     uint64                     `json:"fileLength,omitempty"`
	Filelength     uint64                     `json:"filelength,omitempty"`
	Mime           string                     `json:"mime,omitempty"`
	MimeType       string                     `json:"mimeType,omitempty"`
	Seconds        uint32                     `json:"seconds,omitempty"`
	TypingDuration int                        `json:"typingDuration,omitempty"`
	MediaType      string                     `json:"mediaType,omitempty"`
	Url            string                     `json:"url,omitempty"`
	Content        string                     `json:"content,omitempty"`
	Poll           *whatsapp.WhatsappPoll     `json:"poll,omitempty"`
	Location       *whatsapp.WhatsappLocation `json:"location,omitempty"`
	Contact        *whatsapp.WhatsappContact  `json:"contact,omitempty"`
}

type messageMutationCommandData struct {
	Token     string `json:"token"`
	MessageID string `json:"messageId,omitempty"`
	MessageId string `json:"messageid,omitempty"`
	Content   string `json:"content,omitempty"`
}

type chatArchiveCommandData struct {
	Token   string `json:"token"`
	ChatID  string `json:"chatId,omitempty"`
	ChatId  string `json:"chatid,omitempty"`
	Archive bool   `json:"archive"`
}

type chatPresenceCommandData struct {
	Token    string                            `json:"token"`
	ChatID   string                            `json:"chatId,omitempty"`
	ChatId   string                            `json:"chatid,omitempty"`
	Type     whatsapp.WhatsappChatPresenceType `json:"type"`
	Duration uint                              `json:"duration,omitempty"`
}

func (hub *Hub) registerDefaultCommands() {
	hub.commands["ping"] = hub.handlePing
	hub.commands["subscribe"] = hub.handleSubscribe
	hub.commands["unsubscribe"] = hub.handleUnsubscribe
	hub.commands["server.enable"] = hub.handleServerEnable
	hub.commands["server.disable"] = hub.handleServerDisable
	hub.commands["message.send"] = hub.handleMessageSend
	hub.commands["message.edit"] = hub.handleMessageEdit
	hub.commands["message.revoke"] = hub.handleMessageRevoke
	hub.commands["chat.archive"] = hub.handleChatArchive
	hub.commands["chat.presence"] = hub.handleChatPresence
}

func (hub *Hub) handleCommand(client *Client, command ClientCommand) {
	handler, ok := hub.commands[strings.TrimSpace(command.Command)]
	if !ok {
		hub.sendResponse(client, command, nil, fmt.Errorf("unsupported command: %s", command.Command))
		return
	}

	data, err := handler(client, command)
	hub.sendResponse(client, command, data, err)
}

func (hub *Hub) handlePing(client *Client, command ClientCommand) (interface{}, error) {
	return PingResponsePayload{
		ConnectionID:  client.id,
		User:          client.user.Username,
		Subscriptions: hub.getSubscriptions(client),
	}, nil
}

func (hub *Hub) handleSubscribe(client *Client, command ClientCommand) (interface{}, error) {
	var data subscriptionCommandData
	if err := decodeCommandData(command.Data, &data); err != nil {
		return nil, err
	}

	tokens := normalizeSubscriptionTokens(data)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("missing token/topics")
	}

	subscribed := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if _, err := getOwnedServerRecord(client.user, token); err != nil {
			return nil, err
		}

		hub.SubscribeServer(client, token)
		subscribed = append(subscribed, serverTopic(token))
	}

	sort.Strings(subscribed)
	return SubscriptionResponsePayload{
		Subscriptions: subscribed,
	}, nil
}

func (hub *Hub) handleUnsubscribe(client *Client, command ClientCommand) (interface{}, error) {
	var data subscriptionCommandData
	if err := decodeCommandData(command.Data, &data); err != nil {
		return nil, err
	}

	tokens := normalizeSubscriptionTokens(data)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("missing token/topics")
	}

	removed := make([]string, 0, len(tokens))
	for _, token := range tokens {
		hub.UnsubscribeServer(client, token)
		removed = append(removed, serverTopic(token))
	}

	sort.Strings(removed)
	return SubscriptionResponsePayload{
		Subscriptions: hub.getSubscriptions(client),
		Removed:       removed,
	}, nil
}

func (hub *Hub) handleServerEnable(client *Client, command ClientCommand) (interface{}, error) {
	server, err := getOwnedLiveServerForCommand(client.user, command.Data, true)
	if err != nil {
		return nil, err
	}

	if err := server.Start(); err != nil {
		return nil, err
	}

	return buildServerResult(server), nil
}

func (hub *Hub) handleServerDisable(client *Client, command ClientCommand) (interface{}, error) {
	server, err := getOwnedLiveServerForCommand(client.user, command.Data, false)
	if err != nil {
		return nil, err
	}

	if err := server.Stop("disabled via cable"); err != nil {
		return nil, err
	}

	return buildServerResult(server), nil
}

func (hub *Hub) handleMessageSend(client *Client, command ClientCommand) (interface{}, error) {
	var data sendCommandData
	if err := decodeCommandData(command.Data, &data); err != nil {
		return nil, err
	}

	token := normalizeToken(data.Token)
	if token == "" {
		return nil, fmt.Errorf("missing token")
	}

	if _, err := getOwnedServerRecord(client.user, token); err != nil {
		return nil, err
	}

	server, err := models.WhatsappService.FindByToken(token)
	if err != nil {
		return nil, err
	}

	return sendMessageThroughServer(server, command.ID, &data)
}

func (hub *Hub) handleMessageEdit(client *Client, command ClientCommand) (interface{}, error) {
	var data messageMutationCommandData
	if err := decodeCommandData(command.Data, &data); err != nil {
		return nil, err
	}

	server, err := getOwnedReadyServerForToken(client.user, data.Token)
	if err != nil {
		return nil, err
	}

	messageID := normalizeMessageID(data.MessageID, data.MessageId)
	if messageID == "" {
		return nil, fmt.Errorf("missing messageId")
	}

	content := strings.TrimSpace(data.Content)
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	if err := server.Edit(messageID, content); err != nil {
		return nil, err
	}

	return MessageMutationResponsePayload{
		Token:     server.Token,
		WID:       server.GetWId(),
		MessageID: messageID,
		Action:    "edited",
	}, nil
}

func (hub *Hub) handleMessageRevoke(client *Client, command ClientCommand) (interface{}, error) {
	var data messageMutationCommandData
	if err := decodeCommandData(command.Data, &data); err != nil {
		return nil, err
	}

	server, err := getOwnedReadyServerForToken(client.user, data.Token)
	if err != nil {
		return nil, err
	}

	messageID := normalizeMessageID(data.MessageID, data.MessageId)
	if messageID == "" {
		return nil, fmt.Errorf("missing messageId")
	}

	if err := server.Revoke(messageID); err != nil {
		return nil, err
	}

	return MessageMutationResponsePayload{
		Token:     server.Token,
		WID:       server.GetWId(),
		MessageID: messageID,
		Action:    "revoked",
	}, nil
}

func (hub *Hub) handleChatArchive(client *Client, command ClientCommand) (interface{}, error) {
	var data chatArchiveCommandData
	if err := decodeCommandData(command.Data, &data); err != nil {
		return nil, err
	}

	server, err := getOwnedReadyServerForToken(client.user, data.Token)
	if err != nil {
		return nil, err
	}

	chatID, err := normalizeChatID(data.ChatID, data.ChatId)
	if err != nil {
		return nil, err
	}

	rawConn, err := server.GetValidConnection()
	if err != nil {
		return nil, err
	}

	conn, ok := rawConn.(*whatsmeow.WhatsmeowConnection)
	if !ok {
		return nil, fmt.Errorf("unsupported connection type for archive")
	}

	if err := whatsmeow.ArchiveChat(conn, chatID, data.Archive); err != nil {
		action := "archive"
		if !data.Archive {
			action = "unarchive"
		}
		return nil, fmt.Errorf("failed to %s chat: %s", action, err.Error())
	}

	action := "archived"
	if !data.Archive {
		action = "unarchived"
	}

	return ChatMutationResponsePayload{
		Token:  server.Token,
		WID:    server.GetWId(),
		ChatID: chatID,
		Action: action,
	}, nil
}

func (hub *Hub) handleChatPresence(client *Client, command ClientCommand) (interface{}, error) {
	var data chatPresenceCommandData
	if err := decodeCommandData(command.Data, &data); err != nil {
		return nil, err
	}

	server, err := getOwnedReadyServerForToken(client.user, data.Token)
	if err != nil {
		return nil, err
	}

	chatID, err := normalizeChatID(data.ChatID, data.ChatId)
	if err != nil {
		return nil, err
	}

	previous := cablePresenceRequests.Cancel(chatID)
	if err := server.SendChatPresence(chatID, data.Type); err != nil {
		return nil, fmt.Errorf("failed to send presence update: %s", err.Error())
	}

	if data.Type != whatsapp.WhatsappChatPresenceTypePaused {
		cablePresenceRequests.Append(chatID, data.Type, data.Duration, server)
	}

	return ChatMutationResponsePayload{
		Token:    server.Token,
		WID:      server.GetWId(),
		ChatID:   chatID,
		Action:   fmt.Sprintf("presence:%s", data.Type.String()),
		Previous: previous,
	}, nil
}

func decodeCommandData(raw json.RawMessage, out interface{}) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

func normalizeSubscriptionTokens(data subscriptionCommandData) []string {
	values := []string{data.Token, data.Topic}
	values = append(values, data.Tokens...)
	values = append(values, data.Topics...)

	seen := map[string]struct{}{}
	tokens := make([]string, 0, len(values))
	for _, value := range values {
		token := normalizeSubscriptionInput(value)
		if token == "" {
			continue
		}

		if _, exists := seen[token]; exists {
			continue
		}

		seen[token] = struct{}{}
		tokens = append(tokens, token)
	}

	sort.Strings(tokens)
	return tokens
}

func normalizeSubscriptionInput(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "server:")
	return normalizeToken(value)
}

func getOwnedLiveServerForCommand(user *models.QpUser, raw json.RawMessage, createIfMissing bool) (*models.QpWhatsappServer, error) {
	var data serverTokenCommandData
	if err := decodeCommandData(raw, &data); err != nil {
		return nil, err
	}

	token := normalizeToken(data.Token)
	if token == "" {
		return nil, fmt.Errorf("missing token")
	}

	if _, err := getOwnedServerRecord(user, token); err != nil {
		return nil, err
	}

	if createIfMissing {
		return models.WhatsappService.GetOrCreateServerFromToken(token)
	}

	return models.WhatsappService.FindByToken(token)
}

func getOwnedReadyServerForToken(user *models.QpUser, token string) (*models.QpWhatsappServer, error) {
	token = normalizeToken(token)
	if token == "" {
		return nil, fmt.Errorf("missing token")
	}

	if _, err := getOwnedServerRecord(user, token); err != nil {
		return nil, err
	}

	server, err := models.WhatsappService.FindByToken(token)
	if err != nil {
		return nil, err
	}

	if status := server.GetStatus(); status != whatsapp.Ready {
		return nil, fmt.Errorf("server not ready: %s", status.String())
	}

	if server.Handler == nil {
		return nil, fmt.Errorf("handlers not attached")
	}

	if _, err := server.GetValidConnection(); err != nil {
		return nil, err
	}

	return server, nil
}

func getOwnedServerRecord(user *models.QpUser, token string) (*models.QpServer, error) {
	server, err := models.WhatsappService.DB.Servers.FindByToken(token)
	if err != nil {
		return nil, err
	}

	if server.User != user.Username {
		return nil, fmt.Errorf("server token not owned by user")
	}

	return server, nil
}

func buildServerResult(server *models.QpWhatsappServer) ServerStatePayload {
	return ServerStatePayload{
		Token:    server.Token,
		User:     server.User,
		WID:      server.GetWId(),
		State:    server.GetState().String(),
		Verified: server.Verified,
	}
}

func sendMessageThroughServer(server *models.QpWhatsappServer, commandID string, data *sendCommandData) (interface{}, error) {
	request := newSendMessageRequest(data, commandID)

	if err := request.EnsureValidChatID(); err != nil {
		return nil, err
	}

	if err := request.BuildContent(); err != nil {
		return nil, err
	}

	attachment := request.ToWhatsAppAttachment().Attach
	message, err := request.ToWhatsAppMessage()
	if err != nil {
		return nil, err
	}

	if request.Poll == nil && request.Location == nil && request.Contact == nil && attachment == nil && message.Text == "" {
		return nil, fmt.Errorf("text not found, do not send empty messages")
	}

	if attachment != nil {
		message.Attachment = attachment
		message.Type = whatsapp.GetMessageType(attachment)
	}

	if status := server.GetStatus(); status != whatsapp.Ready {
		return nil, fmt.Errorf("server not ready: %s", status.String())
	}

	if strings.Contains(message.Chat.Id, whatsapp.WHATSAPP_SERVERDOMAIN_LID_SUFFIX) {
		phone, err := server.GetPhoneFromLID(message.Chat.Id)
		if err == nil && phone != "" {
			message.Chat.Id = whatsapp.PhoneToWid(phone)
		}
	} else if !strings.Contains(message.Chat.Id, whatsapp.WHATSAPP_SERVERDOMAIN_USER_SUFFIX) {
		message.Chat.Id = whatsapp.PhoneToWid(message.Chat.Id)
	}

	response, err := server.SendMessage(message)
	if err != nil {
		return nil, err
	}

	return SendMessageResponsePayload{
		ID:      response.GetId(),
		ChatID:  message.Chat.Id,
		TrackID: message.TrackId,
		WID:     server.GetWId(),
		Token:   server.Token,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func firstNonZero(values ...uint64) uint64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func normalizeMessageID(values ...string) string {
	return firstNonEmpty(values...)
}

func normalizeChatID(values ...string) (string, error) {
	chatID := firstNonEmpty(values...)
	if chatID == "" {
		return "", fmt.Errorf("chatId is required")
	}

	return whatsapp.FormatEndpoint(chatID)
}
