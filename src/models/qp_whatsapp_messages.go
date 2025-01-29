package models

import (
	"fmt"
	"strings"
	"time"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

const DEFAULTEXPIRATION time.Duration = time.Duration(124 * time.Hour)

type QpWhatsappMessages struct {
	QpCache

	statuses QpCache
}

func GetCacheExpiration() time.Time {
	var duration time.Duration
	daysFromEnvironment := ENV.CacheDays()
	if daysFromEnvironment > 0 {
		duration = time.Duration(daysFromEnvironment*24) * time.Hour
	} else {
		duration = DEFAULTEXPIRATION
	}

	return time.Now().Add(duration)
}

//#region MESSAGES

func (source *QpWhatsappMessages) Append(value *whatsapp.WhatsappMessage, from string) bool {

	// ensure that is an uppercase string before save
	normalizedId := strings.ToUpper(value.Id)

	expiration := GetCacheExpiration()
	item := QpCacheItem{normalizedId, value, expiration}
	return source.SetCacheItem(item, "message-"+from)
}

func (source *QpWhatsappMessages) GetSlice() (items []*whatsapp.WhatsappMessage) {
	for _, item := range source.GetSliceOfCachedItems() {
		items = append(items, item.Value.(*whatsapp.WhatsappMessage))
	}

	return
}

func (source *QpWhatsappMessages) GetById(id string) (msg *whatsapp.WhatsappMessage, err error) {

	// ensure that is an uppercase string before save
	normalizedId := strings.ToUpper(id)

	cached, found := source.GetAny(normalizedId)
	if !found {
		err = fmt.Errorf("message not present on cache, id: %s", normalizedId)
		return
	}

	msg, ok := cached.(*whatsapp.WhatsappMessage)
	if !ok || msg == nil {
		err = fmt.Errorf("message is corrupted, id: %s", normalizedId)
	}
	return
}

// Returns current cached messages, based on a time filter
func (source *QpWhatsappMessages) GetByTime(timestamp time.Time) (messages []*whatsapp.WhatsappMessage) {

	for _, item := range source.GetSlice() {
		if item.Timestamp.After(timestamp) {
			messages = append(messages, item)
		}
	}

	return
}

// Returns the first in time message stored in cache, used for resync history with message services like whatsapp
func (source *QpWhatsappMessages) GetLeading() (message *whatsapp.WhatsappMessage) {

	now := time.Now()
	for _, item := range source.GetSlice() {
		if !item.Timestamp.IsZero() && item.Timestamp.Before(now) {
			now = item.Timestamp
			message = item
		}
	}

	return
}

func (source *QpWhatsappMessages) GetByPrefix(id string) (messages []*whatsapp.WhatsappMessage) {

	// ensure that is an uppercase string before save
	normalizedId := strings.ToUpper(id)

	for _, item := range source.GetSlice() {
		if strings.HasPrefix(item.Id, normalizedId) {
			messages = append(messages, item)
		}
	}

	return
}

//#endregion
//#region STATUS

func (source *QpWhatsappMessages) SetStatusById(id string, status whatsapp.WhatsappMessageStatus) {

	// ensure that is an uppercase string before save
	normalizedId := strings.ToUpper(id)

	expiration := GetCacheExpiration()
	item := QpCacheItem{normalizedId, status, expiration}
	source.statuses.SetCacheItem(item, "status")
}

func (source *QpWhatsappMessages) GetStatusById(id string) (status whatsapp.WhatsappMessageStatus) {

	// ensure that is an uppercase string before save
	normalizedId := strings.ToUpper(id)

	cached, found := source.statuses.GetAny(normalizedId)
	if found {
		status = cached.(whatsapp.WhatsappMessageStatus)
	}

	return
}

func (source *QpWhatsappMessages) MessageStatusUpdate(id string, status whatsapp.WhatsappMessageStatus) bool {
	cached := source.GetStatusById(id)
	if cached.Uint32() < status.Uint32() {
		source.SetStatusById(id, status)

		msg, err := source.GetById(id)
		if err == nil {
			msg.Status = status
		}

		return true
	}

	return false
}

//#endregion
