package models

import (
	"strconv"
	"strings"
	"time"
)

// returning []QPMessageV1
func GetMessagesFromBotV2(source QPBot, timestamp string) (messages []QpMessageV2, err error) {

	server, err := GetServerFromBot(source)
	if err != nil {
		return
	}

	searchTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		if len(timestamp) > 0 {
			return
		} else {
			searchTimestamp = 0
		}
	}

	searchTime := time.Unix(searchTimestamp, 0)
	messages = GetMessagesFromServerV2(server, searchTime)
	return
}

func ToQpServerV2(source *QpServer) (destination *QpServerV2) {
	destination = &QpServerV2{
		ID:              source.Wid,
		Verified:        source.Verified,
		Token:           source.Token,
		UserID:          source.User,
		Devel:           source.Devel,
		HandleGroups:    source.Groups.ToBoolean(false),
		HandleBroadcast: source.Broadcasts.ToBoolean(false),
		UpdatedAt:       source.Timestamp.String(),
		Version:         "multi",
	}

	if !strings.Contains(destination.ID, "@") {
		destination.ID = destination.ID + "@c.us"
	}
	return
}
