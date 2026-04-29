package models

import (
	"fmt"
	"sort"
	"strings"
	"time"

	cache "github.com/nocodeleaks/quepasa/cache"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	logrus "github.com/sirupsen/logrus"
)

const DEFAULTEXPIRATION time.Duration = time.Duration(124 * time.Hour)

type QpWhatsappMessages struct {
	backend cache.MessagesBackend
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

// SetBackend sets the messages backend for this cache instance.
// This is called by the centralized CacheService during application initialization.
func (source *QpWhatsappMessages) SetBackend(backend cache.MessagesBackend) {
	source.backend = backend
}

func (source *QpWhatsappMessages) readRecord(id string) (cache.MessageRecord, bool, error) {
	if source.backend == nil {
		return cache.MessageRecord{}, false, fmt.Errorf("message cache backend not initialized")
	}

	record, found, err := source.backend.Get(strings.ToUpper(id))
	if err != nil {
		return cache.MessageRecord{}, false, err
	}

	if found && !record.ExpiresAt.IsZero() && time.Now().After(record.ExpiresAt) {
		_ = source.backend.Delete(strings.ToUpper(id))
		return cache.MessageRecord{}, false, nil
	}

	return record, found, nil
}

func (source *QpWhatsappMessages) writeRecord(id string, record cache.MessageRecord) bool {
	if source.backend == nil {
		logrus.Errorf("message cache backend not initialized")
		return false
	}

	err := source.backend.Set(strings.ToUpper(id), record)
	if err != nil {
		logrus.Errorf("failed to persist message cache record: %v", err)
		return false
	}

	return true
}

func (source *QpWhatsappMessages) listRecords() []cache.MessageRecordEntry {
	if source.backend == nil {
		logrus.Errorf("message cache backend not initialized")
		return nil
	}

	entries, err := source.backend.List()
	if err != nil {
		logrus.Errorf("failed to list message cache records: %v", err)
		return nil
	}

	active := make([]cache.MessageRecordEntry, 0, len(entries))
	for _, entry := range entries {
		if !entry.Record.ExpiresAt.IsZero() && time.Now().After(entry.Record.ExpiresAt) {
			_ = source.backend.Delete(entry.Key)
			continue
		}
		active = append(active, entry)
	}

	return active
}

func (source *QpWhatsappMessages) Count() uint64 {
	return uint64(len(source.listRecords()))
}

//#region MESSAGES

func (source *QpWhatsappMessages) Append(value *whatsapp.WhatsappMessage, from string) bool {

	// ensure that is an uppercase string before save
	normalizedId := strings.ToUpper(value.Id)
	value.Id = normalizedId

	expiration := GetCacheExpiration()
	previousRecord, found, err := source.readRecord(normalizedId)
	if err != nil {
		logrus.Errorf("failed to read existing message cache record: %v", err)
		return false
	}

	if len(value.Status) == 0 && previousRecord.Message != nil && len(previousRecord.Message.Status) > 0 {
		value.Status = previousRecord.Message.Status
	}

	record := cache.MessageRecord{
		Message:   value,
		ExpiresAt: expiration,
		UpdatedAt: time.Now(),
	}

	valid := true
	if found {
		previous := QpCacheItem{Key: normalizedId, Value: previousRecord.Message, Expiration: previousRecord.ExpiresAt}
		current := QpCacheItem{Key: normalizedId, Value: value, Expiration: expiration}
		valid = ValidateItemBecauseUNOAPIConflict(current, "message-"+from, previous)
	}

	if !source.writeRecord(normalizedId, record) {
		return false
	}

	return valid
}

func (source *QpWhatsappMessages) GetSlice() (items []*whatsapp.WhatsappMessage) {
	for _, entry := range source.listRecords() {
		if entry.Record.Message != nil {
			items = append(items, entry.Record.Message)
		}
	}

	return
}

func (source *QpWhatsappMessages) GetById(id string) (msg *whatsapp.WhatsappMessage, err error) {

	// ensure that is an uppercase string before save
	normalizedId := strings.ToUpper(id)

	record, found, err := source.readRecord(normalizedId)
	if err != nil {
		return nil, err
	}

	if !found || record.Message == nil {
		err = fmt.Errorf("message not present on cache, id: %s", normalizedId)
		return
	}

	msg, ok := record.Message, record.Message != nil
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

func (source *QpWhatsappMessages) CleanUp(max uint64) {
	if max == 0 {
		return
	}

	entries := source.listRecords()
	if uint64(len(entries)) <= max {
		return
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Record.ExpiresAt.Equal(entries[j].Record.ExpiresAt) {
			return entries[i].Key < entries[j].Key
		}
		return entries[i].Record.ExpiresAt.Before(entries[j].Record.ExpiresAt)
	})

	if source.backend == nil {
		logrus.Errorf("message cache backend not initialized")
		return
	}

	amount := uint64(len(entries)) - max
	for i := uint64(0); i < amount; i++ {
		_ = source.backend.Delete(entries[i].Key)
	}
}

//#endregion
//#region STATUS

func (source *QpWhatsappMessages) SetStatusById(id string, status whatsapp.WhatsappMessageStatus) {

	// ensure that is an uppercase string before save
	normalizedId := strings.ToUpper(id)

	record, found, err := source.readRecord(normalizedId)
	if err != nil {
		logrus.Errorf("failed to read message status record: %v", err)
		return
	}

	if !found {
		return
	}

	record.ExpiresAt = GetCacheExpiration()
	record.UpdatedAt = time.Now()
	if record.Message != nil {
		record.Message.Status = status
	} else {
		return
	}

	_ = source.writeRecord(normalizedId, record)
}

func (source *QpWhatsappMessages) GetStatusById(id string) (status whatsapp.WhatsappMessageStatus) {

	// ensure that is an uppercase string before save
	normalizedId := strings.ToUpper(id)

	record, found, err := source.readRecord(normalizedId)
	if err == nil && found && record.Message != nil {
		status = record.Message.Status
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
