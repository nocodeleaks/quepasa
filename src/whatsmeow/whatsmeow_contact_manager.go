package whatsmeow

import (
	"context"
	"errors"
	"fmt"
	"strings"

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "go.mau.fi/whatsmeow"
	types "go.mau.fi/whatsmeow/types"
	events "go.mau.fi/whatsmeow/types/events"
)

// Compile-time interface check
var _ whatsapp.WhatsappContactManagerInterface = (*WhatsmeowContactManager)(nil)

// WhatsmeowContactManager handles all contact-related operations for WhatsmeowConnection
type WhatsmeowContactManager struct {
	*WhatsmeowConnection                       // embedded connection for direct access
	maps                 *WhatsmeowContactMaps // global contact mappings singleton
}

// NewWhatsmeowContactManager creates a new WhatsmeowContactManager instance
func NewWhatsmeowContactManager(conn *WhatsmeowConnection) *WhatsmeowContactManager {
	return &WhatsmeowContactManager{
		WhatsmeowConnection: conn,
		maps:                GetGlobalContactMaps(), // Use singleton instance
	}
}

// GetContacts returns all contacts from WhatsApp
func (cm *WhatsmeowContactManager) GetContacts() (chats []whatsapp.WhatsappChat, err error) {
	if cm.Client == nil {
		err = errors.New("invalid client")
		return chats, err
	}

	if cm.Client.Store == nil {
		err = errors.New("invalid store")
		return chats, err
	}

	// Delegate to shared helper function
	return GetContactsFromDevice(cm.Client.Store)
}

// IsOnWhatsApp checks if phone numbers are registered on WhatsApp.
//
// WARNING: This method performs a live query against WhatsApp servers.
// Results are cached in-memory (per session singleton) to avoid repeated
// network calls for the same phone number. Bypassing the cache or calling
// with many distinct numbers in a short period may trigger WhatsApp's
// anti-abuse detection and result in account banning.
func (cm *WhatsmeowContactManager) IsOnWhatsApp(phones ...string) (registered []string, err error) {
	var uncached []string

	// Return cached results immediately; collect phones that still need lookup
	for _, phone := range phones {
		if jid, found := cm.maps.GetIsOnWhatsAppCache(phone); found {
			if jid != "" {
				registered = append(registered, jid)
			}
		} else {
			uncached = append(uncached, phone)
		}
	}

	if len(uncached) == 0 {
		return
	}

	// Live query only for phones not yet cached
	results, err := cm.Client.IsOnWhatsApp(context.Background(), uncached)
	if err != nil {
		return
	}

	// Build a set of which uncached phones got a positive result
	resolved := make(map[string]string, len(results))
	for _, result := range results {
		if result.IsIn {
			resolved[result.Query] = result.JID.String()
			registered = append(registered, result.JID.String())
		}
	}

	// Persist results in cache (including negatives, to avoid future lookups)
	for _, phone := range uncached {
		cm.maps.SetIsOnWhatsAppCache(phone, resolved[phone])
	}

	return
}

// GetProfilePicture gets profile picture information
func (cm *WhatsmeowContactManager) GetProfilePicture(wid string, knowingId string) (picture *whatsapp.WhatsappProfilePicture, err error) {
	jid, err := types.ParseJID(wid)
	if err != nil {
		return
	}

	params := &whatsmeow.GetProfilePictureParams{}
	params.ExistingID = knowingId
	params.Preview = false

	pictureInfo, err := cm.Client.GetProfilePictureInfo(context.Background(), jid, params)
	if err != nil {
		return
	}

	if pictureInfo != nil {
		picture = &whatsapp.WhatsappProfilePicture{
			Id:   pictureInfo.ID,
			Type: pictureInfo.Type,
			Url:  pictureInfo.URL,
		}
	}
	return
}

// GetLIDFromPhone returns the @lid for a given phone number
// IMPORTANT: This method accepts phone numbers in E164 format (with +) or without +
// and always normalizes them by removing the + before creating JIDs for WhatsApp API calls
func (cm *WhatsmeowContactManager) GetLIDFromPhone(phone string) (string, error) {
	// Safety check: verify if ContactManager is not nil
	if cm == nil {
		return "", fmt.Errorf("contact manager is nil")
	}

	// Safety check: verify if maps is not nil
	if cm.maps == nil {
		return "", fmt.Errorf("contact maps is nil")
	}

	logger := cm.GetLogger()

	normalized := strings.TrimSpace(phone)
	normalized = strings.TrimPrefix(normalized, "+") // Remove leading + if present

	// First, check maps for existing mapping - this should be the very first check
	if cachedLID, exists := cm.maps.GetLIDFromPhoneMap(normalized); exists {
		logger.Debugf("Found LID in maps for phone %s: %s", phone, cachedLID)
		return cachedLID, nil
	}

	if cm.Client == nil {
		return "", fmt.Errorf("client not defined")
	}

	if cm.Client.Store == nil {
		return "", fmt.Errorf("store not defined")
	}

	logger.Debugf("Phone %s not found in maps, querying database", normalized)

	// Parse the phone number to JID format
	phoneJID := types.JID{
		User:   normalized,
		Server: whatsapp.WHATSAPP_SERVERDOMAIN_USER,
	}

	// try to get the LID from local store
	lidJID, err := cm.Client.Store.LIDs.GetLIDForPN(context.Background(), phoneJID)
	if err == nil && !lidJID.IsEmpty() {
		lid := lidJID.ToNonAD().String()
		logger.Debugf("LID found in database for phone %s: %s", phone, lid)

		// Caching successful mapping for future use
		cm.maps.SetLIDFromPhoneMap(normalized, lid)
		logger.Debugf("Phone->LID mapping cached: %s -> %s", normalized, lid)

		// TEMPORARY WORKAROUND: Brazilian mobile phones exist in two variants —
		// 8-digit (legacy, pre-migration) and 9-digit (modern, with extra leading 9 after DDD).
		// WhatsApp has not standardized which variant is canonical for each account.
		// When we successfully resolve a mapping for one variant, we persist the other variant
		// in Store.LIDs so that whatsmeow's SendMessage PN→LID conversion path (triggered
		// when LIDMigrationTimestamp > 0) can resolve either form without a round-trip to
		// the WhatsApp server. AllDDDs variants are used here (no DDD > 30 restriction)
		// because persisting an extra mapping that is never looked up is harmless.
		// Remove this block once WhatsApp enforces a single canonical phone format.
		if variantPhone, verr := library.AddDigit9BRAllDDDs("+" + normalized); verr == nil {
			variantNormalized := strings.TrimPrefix(variantPhone, "+")
			variantJID := types.JID{User: variantNormalized, Server: whatsapp.WHATSAPP_SERVERDOMAIN_USER}
			if perr := cm.Client.Store.LIDs.PutLIDMapping(context.Background(), lidJID, variantJID); perr == nil {
				logger.Debugf("BR digit-9 variant persisted in Store.LIDs: %s -> %s", variantNormalized, lid)
				cm.maps.SetLIDFromPhoneMap(variantNormalized, lid)
			} else {
				logger.Warnf("BR digit-9 variant Store.LIDs write failed for %s: %v", variantNormalized, perr)
			}
		} else if variantPhone, verr := library.RemoveDigit9BRAllDDDs("+" + normalized); verr == nil {
			variantNormalized := strings.TrimPrefix(variantPhone, "+")
			variantJID := types.JID{User: variantNormalized, Server: whatsapp.WHATSAPP_SERVERDOMAIN_USER}
			if perr := cm.Client.Store.LIDs.PutLIDMapping(context.Background(), lidJID, variantJID); perr == nil {
				logger.Debugf("BR digit-9 variant persisted in Store.LIDs: %s -> %s", variantNormalized, lid)
				cm.maps.SetLIDFromPhoneMap(variantNormalized, lid)
			} else {
				logger.Warnf("BR digit-9 variant Store.LIDs write failed for %s: %v", variantNormalized, perr)
			}
		}

		return lid, nil
	}

	// TEMPORARY WORKAROUND: direct lookup failed — try the BR digit-9 variant before giving up.
	// The phone in the store may be the opposite form (e.g. stored as 9-digit, queried as 8-digit).
	// All Brazilian DDDs are tried since the store mapping may exist for any DDD.
	var variantNormalizedFallback string
	if vp, verr := library.AddDigit9BRAllDDDs("+" + normalized); verr == nil {
		variantNormalizedFallback = strings.TrimPrefix(vp, "+")
	} else if vp, verr := library.RemoveDigit9BRAllDDDs("+" + normalized); verr == nil {
		variantNormalizedFallback = strings.TrimPrefix(vp, "+")
	}

	if variantNormalizedFallback != "" {
		// Check in-memory map for variant first
		if cachedLID, exists := cm.maps.GetLIDFromPhoneMap(variantNormalizedFallback); exists {
			logger.Debugf("LID found in maps via BR digit-9 variant for phone %s: %s", phone, cachedLID)
			cm.maps.SetLIDFromPhoneMap(normalized, cachedLID)
			return cachedLID, nil
		}

		variantJIDFallback := types.JID{User: variantNormalizedFallback, Server: whatsapp.WHATSAPP_SERVERDOMAIN_USER}
		if lidJID2, err2 := cm.Client.Store.LIDs.GetLIDForPN(context.Background(), variantJIDFallback); err2 == nil && !lidJID2.IsEmpty() {
			lid := lidJID2.ToNonAD().String()
			logger.Debugf("LID found in database via BR digit-9 variant %s: %s", variantNormalizedFallback, lid)

			// Cache both the original and the variant so future lookups skip DB
			cm.maps.SetLIDFromPhoneMap(normalized, lid)
			cm.maps.SetLIDFromPhoneMap(variantNormalizedFallback, lid)

			// Also persist original form in Store.LIDs so whatsmeow's internal path finds it
			originalJID := types.JID{User: normalized, Server: whatsapp.WHATSAPP_SERVERDOMAIN_USER}
			if perr := cm.Client.Store.LIDs.PutLIDMapping(context.Background(), lidJID2, originalJID); perr != nil {
				logger.Warnf("BR digit-9 fallback: Store.LIDs write failed for original %s: %v", normalized, perr)
			}

			return lid, nil
		}
	}

	logger.Debugf("No LID mapping found for phone %s", normalized)
	return "", nil
}

// GetPhoneFromLID returns the phone number for a given @lid
func (cm *WhatsmeowContactManager) GetPhoneFromLID(lid string) (string, error) {
	// Safety check: verify if ContactManager is not nil
	if cm == nil {
		return "", fmt.Errorf("contact manager is nil")
	}

	// Safety check: verify if maps is not nil
	if cm.maps == nil {
		return "", fmt.Errorf("contact maps is nil")
	}

	logger := cm.GetLogger()

	// First, check maps for existing mapping - this should be the very first check
	if cachedPhone, exists := cm.maps.GetPhoneFromLIDMap(lid); exists {
		logger.Debugf("Found phone in maps for LID %s: %s", lid, cachedPhone)
		return cachedPhone, nil
	}

	if cm.Client == nil {
		return "", fmt.Errorf("client not defined")
	}

	if cm.Client.Store == nil {
		return "", fmt.Errorf("store not defined")
	}

	logger.Debugf("LID %s not found in maps, querying database", lid)

	// Parse the LID to JID format
	lidJID, err := types.ParseJID(lid)
	if err != nil {
		return "", fmt.Errorf("invalid LID format: %v", err)
	}

	// Get the corresponding phone number from local store
	phoneJID, err := cm.Client.Store.LIDs.GetPNForLID(context.Background(), lidJID)
	if err != nil {
		return "", fmt.Errorf("failed to get phone for LID %s: %v", lid, err)
	}

	if phoneJID.IsEmpty() {
		return "", fmt.Errorf("no phone found for LID %s", lid)
	}

	phone := phoneJID.User
	logger.Debugf("Phone found in database for LID %s: %s", lid, phone)

	// Store successful mapping for future use
	cm.maps.SetPhoneFromLIDMap(lid, phone)
	logger.Debugf("LID->Phone mapping stored: %s -> %s", lid, phone)

	// TEMPORARY WORKAROUND: Brazilian mobile phones exist in two variants —
	// 8-digit (legacy, pre-migration) and 9-digit (modern, with extra leading 9 after DDD).
	// When we resolve a phone from a LID, we persist the alternate variant in Store.LIDs
	// so that subsequent lookups (and whatsmeow's internal PN→LID path) find either form.
	// AllDDDs variants are used here — no DDD > 30 restriction since storing an extra
	// mapping that is never looked up is harmless. For phone-only sends keep using the
	// restricted (DDD > 30) helpers to avoid wrong digit manipulation.
	// Remove this block once WhatsApp enforces a single canonical phone format.
	if variantPhone, verr := library.AddDigit9BRAllDDDs("+" + phone); verr == nil {
		variantNormalized := strings.TrimPrefix(variantPhone, "+")
		variantJID := types.JID{User: variantNormalized, Server: whatsapp.WHATSAPP_SERVERDOMAIN_USER}
		if perr := cm.Client.Store.LIDs.PutLIDMapping(context.Background(), lidJID, variantJID); perr == nil {
			logger.Debugf("BR digit-9 variant persisted in Store.LIDs (reverse): %s -> %s", lid, variantNormalized)
			cm.maps.SetLIDFromPhoneMap(variantNormalized, lid)
		} else {
			logger.Warnf("BR digit-9 variant Store.LIDs write failed (reverse) for %s: %v", variantNormalized, perr)
		}
	} else if variantPhone, verr := library.RemoveDigit9BRAllDDDs("+" + phone); verr == nil {
		variantNormalized := strings.TrimPrefix(variantPhone, "+")
		variantJID := types.JID{User: variantNormalized, Server: whatsapp.WHATSAPP_SERVERDOMAIN_USER}
		if perr := cm.Client.Store.LIDs.PutLIDMapping(context.Background(), lidJID, variantJID); perr == nil {
			logger.Debugf("BR digit-9 variant persisted in Store.LIDs (reverse): %s -> %s", lid, variantNormalized)
			cm.maps.SetLIDFromPhoneMap(variantNormalized, lid)
		} else {
			logger.Warnf("BR digit-9 variant Store.LIDs write failed (reverse) for %s: %v", variantNormalized, perr)
		}
	}

	return phone, nil
}

// GetUserInfo retrieves comprehensive user information for given JIDs
func (cm *WhatsmeowContactManager) GetUserInfo(jids []string) ([]interface{}, error) {
	if cm.Client == nil {
		return nil, fmt.Errorf("client not defined")
	}

	if cm.Client.Store == nil {
		return nil, fmt.Errorf("store not defined")
	}

	// Convert string JIDs to types.JID
	var parsedJIDs []types.JID
	for _, jidStr := range jids {
		// Check if it's a phone number (no @ symbol) and validate E164 format
		if !strings.Contains(jidStr, "@") {
			// This is a phone number, validate and format to E164
			validPhone, err := whatsapp.GetPhoneIfValid(jidStr)
			if err != nil {
				return nil, fmt.Errorf("invalid phone number format for %s: %v (must be E164 format starting with +)", jidStr, err)
			}

			// Remove the + from E164 format for JID creation
			phoneNumber := strings.TrimPrefix(validPhone, "+")
			jid := types.JID{
				User:   phoneNumber,
				Server: whatsapp.WHATSAPP_SERVERDOMAIN_USER,
			}
			parsedJIDs = append(parsedJIDs, jid)
		} else {
			// This is already a JID, parse normally
			jid, err := types.ParseJID(jidStr)
			if err != nil {
				return nil, fmt.Errorf("invalid JID format for %s: %v", jidStr, err)
			}
			parsedJIDs = append(parsedJIDs, jid)
		}
	}

	// Get user info from WhatsApp - this returns a map[types.JID]types.UserInfo
	userInfoMap, err := cm.Client.GetUserInfo(context.Background(), parsedJIDs)
	logentry := cm.GetLogger()
	logentry.Debugf("GetUserInfo for JIDs: %v, result: %v", parsedJIDs, userInfoMap)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}

	// Convert map to interface array for generic return type
	result := make([]interface{}, 0, len(userInfoMap))
	for jid, info := range userInfoMap {
		// Get contact info from local store - try both JID and corresponding phone/LID
		contactInfo, contactErr := cm.Client.Store.Contacts.GetContact(context.TODO(), jid)

		// Get LID/Phone mapping information
		var lid, phoneNumber string
		var phoneJID types.JID

		if strings.Contains(jid.String(), whatsapp.WHATSAPP_SERVERDOMAIN_LID_SUFFIX) {
			// This is a LID, try to get corresponding phone
			lid = jid.ToNonAD().String()
			pnJID, err := cm.Client.Store.LIDs.GetPNForLID(context.TODO(), jid)
			if err == nil && !pnJID.IsEmpty() {
				phoneNumber = pnJID.User
				phoneJID = pnJID

				// If we didn't get contact info from LID, try with phone JID
				if contactErr != nil {
					contactInfo, contactErr = cm.Client.Store.Contacts.GetContact(context.TODO(), phoneJID)
				}
			}
		} else {
			// This is a phone number JID, try to get corresponding LID
			phoneNumber = jid.User
			lidJID, err := cm.Client.Store.LIDs.GetLIDForPN(context.TODO(), jid)
			if err == nil && !lidJID.IsEmpty() {
				lid = lidJID.ToNonAD().String()

				// If we didn't get contact info from phone JID, try with LID
				if contactErr != nil {
					contactInfo, contactErr = cm.Client.Store.Contacts.GetContact(context.TODO(), lidJID)
				}
			}
		}

		// Format phone to E164 if available
		var phoneE164 string
		if phoneNumber != "" {
			if phone, err := whatsapp.GetPhoneIfValid(phoneNumber); err == nil {
				phoneE164 = phone
			}
		}

		// Determine the best display name
		var displayName string
		if contactErr == nil {
			if contactInfo.FullName != "" {
				displayName = contactInfo.FullName
			} else if contactInfo.BusinessName != "" {
				displayName = contactInfo.BusinessName
			} else if contactInfo.PushName != "" {
				displayName = contactInfo.PushName
			}
		}

		// If no local contact name, use verified name from user info
		if displayName == "" && info.VerifiedName != nil {
			displayName = info.VerifiedName.Details.GetVerifiedName()
		}

		// Check if we have meaningful contact information
		hasContactInfo := contactErr == nil && (contactInfo.FullName != "" || contactInfo.BusinessName != "" || contactInfo.PushName != "")
		hasVerifiedName := info.VerifiedName != nil && info.VerifiedName.Details.GetVerifiedName() != ""
		hasStatus := info.Status != ""
		hasPictureID := info.PictureID != ""
		hasDevices := len(info.Devices) > 0
		hasLID := lid != ""

		// Only include contacts that have meaningful information beyond just phone/JID
		if !hasContactInfo && !hasVerifiedName && !hasStatus && !hasPictureID && !hasDevices && !hasLID {
			logentry.Debugf("Skipping contact %s - no meaningful information possible non whatsapp number", jid.String())
			continue
		}

		// Create a comprehensive response with omitempty support
		userInfoResponse := WhatsmeowUserInfoResponse{
			JID:          jid.String(),
			LID:          lid,
			Phone:        phoneNumber,
			PhoneE164:    phoneE164,
			Status:       info.Status,
			PictureID:    info.PictureID,
			Devices:      info.Devices,
			VerifiedName: info.VerifiedName,
			DisplayName:  displayName,
		}

		// Add contact-specific information if available
		if contactErr == nil {
			userInfoResponse.FullName = contactInfo.FullName
			userInfoResponse.BusinessName = contactInfo.BusinessName
			userInfoResponse.PushName = contactInfo.PushName
		}

		result = append(result, userInfoResponse)
	}

	return result, nil
}

// BlockContact blocks a contact by their WID/JID so they cannot send messages to this account.
func (cm *WhatsmeowContactManager) BlockContact(wid string) error {
	if cm.Client == nil {
		return fmt.Errorf("client not defined")
	}

	jid, err := types.ParseJID(wid)
	if err != nil {
		return fmt.Errorf("invalid contact id: %w", err)
	}

	_, err = cm.Client.UpdateBlocklist(context.Background(), jid, events.BlocklistChangeActionBlock)
	return err
}

// UnblockContact removes a block previously placed on a contact.
func (cm *WhatsmeowContactManager) UnblockContact(wid string) error {
	if cm.Client == nil {
		return fmt.Errorf("client not defined")
	}

	jid, err := types.ParseJID(wid)
	if err != nil {
		return fmt.Errorf("invalid contact id: %w", err)
	}

	_, err = cm.Client.UpdateBlocklist(context.Background(), jid, events.BlocklistChangeActionUnblock)
	return err
}

// GetPhoneFromContactId attempts to get phone number from contact Id using available mapping
func (cm *WhatsmeowContactManager) GetPhoneFromContactId(contactId string) (string, error) {
	if strings.Contains(contactId, whatsapp.WHATSAPP_SERVERDOMAIN_USER_SUFFIX) {
		phone, err := whatsapp.GetPhoneIfValid(contactId)
		if err == nil {
			return phone, nil // Return phone if valid
		}
	}

	logentry := cm.GetLogger()
	logentry = logentry.WithField("entry", "WhatsmeowContactManager.GetPhoneFromContactId")

	// Try to get phone from different sources
	if strings.Contains(contactId, whatsapp.WHATSAPP_SERVERDOMAIN_LID_SUFFIX) {

		logentry.Debug("Attempting to get phone from LId")

		// For @lid, try to get the corresponding phone number using contact manager interface
		if retrieved, err := cm.GetPhoneFromLID(contactId); err == nil && len(retrieved) > 0 {
			logentry.Debugf("Retrieved phone from LId mapping: %s", retrieved)

			// Format the phone to E164 if needed
			if phone, err := whatsapp.GetPhoneIfValid(retrieved); err == nil {
				logentry.Debug("Phone formatted to E164")
				return phone, nil
			}
		} else {
			logentry.WithError(err).Error("Failed to get phone from LID mapping")
			return "", err
		}
	}

	// If still not found, return error
	logentry.Infof("Can't find suitable E164 phone for contact Id: %s", contactId)
	return "", fmt.Errorf("no phone found for contact Id: %s", contactId)
}
