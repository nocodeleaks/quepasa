package whatsmeow

import (
	"strings"
	"sync"
)

// WhatsmeowContactMaps provides thread-safe mapping for LID/Phone relationships
type WhatsmeowContactMaps struct {
	mutex      sync.RWMutex
	lidToPhone map[string]string // LID -> Phone mapping
	phoneToLID map[string]string // Phone -> LID mapping
}

var (
	// Global singleton instance
	globalContactMaps *WhatsmeowContactMaps
	// Mutex to ensure thread-safe singleton initialization
	once sync.Once
)

// GetGlobalContactMaps returns the singleton instance of contact maps
func GetGlobalContactMaps() *WhatsmeowContactMaps {
	once.Do(func() {
		globalContactMaps = &WhatsmeowContactMaps{
			lidToPhone: make(map[string]string),
			phoneToLID: make(map[string]string),
		}
	})
	return globalContactMaps
}

// normalizePhone removes the "+" prefix for storage
func normalizePhone(phone string) string {
	return strings.TrimPrefix(phone, "+")
}

// formatPhone adds the "+" prefix for return
func formatPhone(phone string) string {
	if phone == "" {
		return ""
	}
	if strings.HasPrefix(phone, "+") {
		return phone
	}
	return "+" + phone
}

// GetPhoneFromLIDMap retrieves phone from LID mapping if exists (returns with + prefix)
func (c *WhatsmeowContactMaps) GetPhoneFromLIDMap(lid string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	phone, exists := c.lidToPhone[lid]
	if exists {
		return formatPhone(phone), true
	}
	return "", false
}

// SetPhoneFromLIDMap stores LID->Phone mapping (stores phone without + prefix)
func (c *WhatsmeowContactMaps) SetPhoneFromLIDMap(lid, phone string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Normalize phone for storage (remove + prefix)
	normalizedPhone := normalizePhone(phone)
	c.lidToPhone[lid] = normalizedPhone

	// Also store reverse mapping if phone is not empty
	if len(normalizedPhone) > 0 {
		c.phoneToLID[normalizedPhone] = lid
	}
}

// GetLIDFromPhoneMap retrieves LID from phone mapping if exists (accepts phone with or without + prefix)
func (c *WhatsmeowContactMaps) GetLIDFromPhoneMap(phone string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Normalize phone for lookup (remove + prefix)
	normalizedPhone := normalizePhone(phone)
	lid, exists := c.phoneToLID[normalizedPhone]
	return lid, exists
}

// SetLIDFromPhoneMap stores Phone->LID mapping (stores phone without + prefix)
func (c *WhatsmeowContactMaps) SetLIDFromPhoneMap(phone, lid string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Normalize phone for storage (remove + prefix)
	normalizedPhone := normalizePhone(phone)
	c.phoneToLID[normalizedPhone] = lid

	// Also store reverse mapping if lid is not empty
	if len(lid) > 0 {
		c.lidToPhone[lid] = normalizedPhone
	}
}
