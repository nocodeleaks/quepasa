package controllers

import (
	"sort"
	"strconv"
	"time"

	models "github.com/nocodeleaks/quepasa/models"
)

// Retrieve messages with timestamp parameter
// Sorting then
func GetMessagesToAPIV2(server *models.QpWhatsappServer, timestamp string) (messages []models.QpMessageV2, err error) {

	searchTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		if len(timestamp) > 0 {
			return
		} else {
			err = nil
			searchTimestamp = 0
		}
	}

	searchTime := time.Unix(searchTimestamp, 0)
	messages = models.GetMessagesFromServerV2(server, searchTime)
	sort.Sort(models.ByTimestampV2(messages))
	return
}
