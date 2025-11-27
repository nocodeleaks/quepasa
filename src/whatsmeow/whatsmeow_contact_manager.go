package whatsmeow

import (
	"context"
	"errors"
	"fmt"
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "go.mau.fi/whatsmeow"
	types "go.mau.fi/whatsmeow/types"
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

	contacts, err := cm.Client.Store.Contacts.GetAllContacts(context.TODO())
	if err != nil {
		return chats, err
	}

	// Map to track contacts by phone number
	contactMap := make(map[string]whatsapp.WhatsappChat)

	for jid, info := range contacts {
		title := info.FullName
		if len(title) == 0 {
			title = info.BusinessName
			if len(title) == 0 {
				title = info.PushName
			}
		}

		var phoneNumber string
		var lid string
		var phoneE164 string

		if strings.Contains(jid.String(), whatsapp.WHATSAPP_SERVERDOMAIN_LID_SUFFIX) {
			// For @lid contacts, get the corresponding phone number
			pnJID, err := cm.Client.Store.LIDs.GetPNForLID(context.TODO(), jid)
			if err == nil && !pnJID.IsEmpty() {
				phoneNumber = pnJID.User
				lid = jid.String()
				// Format phone to E164
				if phone, err := whatsapp.GetPhoneIfValid(phoneNumber); err == nil {
					phoneE164 = phone
				}
			} else {
				// If no mapping found, use the LID as unique identifier
				phoneNumber = jid.String()
				lid = ""
			}
		} else {
			// For regular @s.whatsapp.net contacts
			phoneNumber = jid.User
			// Format phone to E164
			if phone, err := whatsapp.GetPhoneIfValid(phoneNumber); err == nil {
				phoneE164 = phone
			}

			// Try to get corresponding LID
			lidJID, err := cm.Client.Store.LIDs.GetLIDForPN(context.TODO(), jid)
			if err == nil && !lidJID.IsEmpty() {
				lid = lidJID.String()
			} else {
				lid = ""
			}
		}

		// Check if contact with this phone number already exists
		existingContact, exists := contactMap[phoneNumber]

		if !exists {
			// First contact with this phone number
			contactMap[phoneNumber] = whatsapp.WhatsappChat{
				Id:    jid.String(),
				LId:   lid,
				Title: title,
				Phone: phoneE164,
			}
		} else {
			// Contact already exists, merge information
			var finalId, finalLId, finalPhone string

			if strings.Contains(jid.String(), whatsapp.WHATSAPP_SERVERDOMAIN_LID_SUFFIX) {
				// Current is @lid, keep existing as Id and use current as Lid
				finalId = existingContact.Id
				finalLId = jid.String()
				finalPhone = existingContact.Phone
				if len(finalPhone) == 0 && len(phoneE164) > 0 {
					finalPhone = phoneE164
				}
			} else {
				// Current is @s.whatsapp.net, use as Id and keep existing LId
				finalId = jid.String()
				finalLId = existingContact.LId
				if len(finalLId) == 0 && len(lid) > 0 {
					finalLId = lid
				}
				finalPhone = phoneE164
				if len(finalPhone) == 0 && len(existingContact.Phone) > 0 {
					finalPhone = existingContact.Phone
				}
			}

			// Keep the best available title
			finalTitle := title
			if len(finalTitle) == 0 && len(existingContact.Title) > 0 {
				finalTitle = existingContact.Title
			}

			contactMap[phoneNumber] = whatsapp.WhatsappChat{
				Id:    finalId,
				LId:   finalLId,
				Title: finalTitle,
				Phone: finalPhone,
			}
		}
	}

	// Convert map to slice
	for _, contact := range contactMap {
		chats = append(chats, contact)
	}

	return chats, nil
}

// IsOnWhatsApp checks if phone numbers are registered on WhatsApp
func (cm *WhatsmeowContactManager) IsOnWhatsApp(phones ...string) (registered []string, err error) {
	results, err := cm.Client.IsOnWhatsApp(context.Background(), phones)
	if err != nil {
		return
	}

	for _, result := range results {
		if result.IsIn {
			registered = append(registered, result.JID.String())
		}
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
		lid := lidJID.String()
		logger.Debugf("LID found in database for phone %s: %s", phone, lid)

		// Caching successful mapping for future use
		cm.maps.SetLIDFromPhoneMap(normalized, lid)
		logger.Debugf("Phone->LID mapping cached: %s -> %s", normalized, lid)

		return lid, nil
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
			lid = jid.String()
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
				lid = lidJID.String()

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
