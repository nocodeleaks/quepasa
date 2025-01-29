package models

import "time"

// returning []QPMessageV1
func GetMessagesFromServerV2(server *QpWhatsappServer, searchTime time.Time) (messages []QpMessageV2) {
	list := server.GetMessages(searchTime)
	for _, item := range list {
		messages = append(messages, ToQpMessageV2(item, server))
	}

	return
}
