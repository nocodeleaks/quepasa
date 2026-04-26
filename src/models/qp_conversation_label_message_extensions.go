package models

import (
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

func CloneAndEnrichMessageForServer(server *QpWhatsappServer, message *whatsapp.WhatsappMessage) *whatsapp.WhatsappMessage {
	if message == nil {
		return nil
	}

	clone := *message
	clone.Chat = message.Chat
	if message.Participant != nil {
		participant := *message.Participant
		clone.Participant = &participant
	}

	applyConversationLabelsToMessage(server, &clone, nil)
	return &clone
}

func CloneAndEnrichMessagesForServer(server *QpWhatsappServer, messages []whatsapp.WhatsappMessage) []whatsapp.WhatsappMessage {
	if len(messages) == 0 {
		return messages
	}

	cloned := make([]whatsapp.WhatsappMessage, len(messages))
	chatIDs := make([]string, 0, len(messages))
	for index, item := range messages {
		cloned[index] = item
		cloned[index].Chat = item.Chat
		if item.Participant != nil {
			participant := *item.Participant
			cloned[index].Participant = &participant
		}

		chatID := strings.TrimSpace(item.Chat.Id)
		if chatID == "" || chatID == whatsapp.WASYSTEMCHAT.Id {
			continue
		}
		chatIDs = append(chatIDs, chatID)
	}

	labelsByChatID := map[string][]*QpConversationLabel{}
	if store, ok := getConversationLabelStore(); ok && server != nil {
		user := strings.TrimSpace(server.GetUser())
		token := strings.TrimSpace(server.Token)
		if user != "" && token != "" {
			if loaded, err := store.FindConversationLabelsMap(token, user, chatIDs); err == nil {
				labelsByChatID = loaded
			}
		}
	}

	for index := range cloned {
		applyConversationLabelsToMessage(server, &cloned[index], labelsByChatID)
	}

	return cloned
}

func applyConversationLabelsToMessage(server *QpWhatsappServer, message *whatsapp.WhatsappMessage, preloaded map[string][]*QpConversationLabel) {
	if server == nil || message == nil {
		return
	}

	chatID := strings.TrimSpace(message.Chat.Id)
	if chatID == "" || chatID == whatsapp.WASYSTEMCHAT.Id {
		return
	}

	var labels []*QpConversationLabel
	if preloaded != nil {
		labels = preloaded[chatID]
	} else if store, ok := getConversationLabelStore(); ok {
		user := strings.TrimSpace(server.GetUser())
		token := strings.TrimSpace(server.Token)
		if user == "" || token == "" {
			return
		}

		loaded, err := store.FindConversationLabels(token, chatID, user)
		if err != nil {
			return
		}
		labels = loaded
	}

	if len(labels) == 0 {
		message.Chat.Labels = nil
		return
	}

	items := make([]whatsapp.WhatsappChatLabel, 0, len(labels))
	for _, label := range labels {
		if label == nil {
			continue
		}
		items = append(items, label.ToWhatsappLabel())
	}
	message.Chat.Labels = items
}

func getConversationLabelStore() (QpDataConversationLabelsInterface, bool) {
	if WhatsappService == nil || WhatsappService.DB == nil || WhatsappService.DB.ConversationLabels == nil {
		return nil, false
	}

	return WhatsappService.DB.ConversationLabels, true
}
