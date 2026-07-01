package whatsmeow

import (
	"context"
	"fmt"
	"strings"

	voip "github.com/nocodeleaks/quepasa/voip"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	types "go.mau.fi/whatsmeow/types"
)

// GetVoIPSectionID returns the stable WhatsApp section identifier used by SIP
// gateways. It is intentionally separate from the QuePasa session token.
func (conn *WhatsmeowConnection) GetVoIPSectionID() string {
	if conn == nil || conn.Client == nil || conn.Client.Store == nil || conn.Client.Store.ID == nil {
		return ""
	}

	jid := conn.Client.Store.ID
	user := strings.TrimSpace(jid.User)
	if user == "" {
		return ""
	}

	if jid.Device > 0 {
		return fmt.Sprintf("%s:%d", user, jid.Device)
	}
	return user
}

// ResolveVoIPCallerInfo returns local QuePasa contact metadata for a WhatsApp
// caller. It intentionally reads only local whatsmeow stores so call setup does
// not depend on live WhatsApp profile/network queries.
func (conn *WhatsmeowConnection) ResolveVoIPCallerInfo(peer types.JID) voip.CallerInfo {
	info := voip.CallerInfo{
		JID:   peer.String(),
		Phone: peer.User,
	}
	if conn == nil || conn.Client == nil || conn.Client.Store == nil {
		return info
	}

	store := conn.Client.Store
	ctx := context.TODO()
	lookupJID := CleanJID(peer)
	if lookupJID.IsEmpty() {
		lookupJID = peer
	}
	info.JID = lookupJID.String()

	var contactErr error
	if store.Contacts != nil {
		var contactInfo types.ContactInfo
		contactInfo, contactErr = store.Contacts.GetContact(ctx, lookupJID)
		if contactErr == nil {
			applyVoIPContactInfo(&info, contactInfo)
		}
	}

	if strings.Contains(lookupJID.String(), whatsapp.WHATSAPP_SERVERDOMAIN_LID_SUFFIX) {
		info.LID = lookupJID.ToNonAD().String()
		if store.LIDs != nil {
			if pnJID, err := store.LIDs.GetPNForLID(ctx, lookupJID); err == nil && !pnJID.IsEmpty() {
				info.Phone = pnJID.User
				info.JID = pnJID.String()
				if contactErr != nil && store.Contacts != nil {
					if pnContact, err := store.Contacts.GetContact(ctx, pnJID); err == nil {
						applyVoIPContactInfo(&info, pnContact)
					}
				}
			}
		}
	} else {
		info.Phone = lookupJID.User
		if store.LIDs != nil {
			if lidJID, err := store.LIDs.GetLIDForPN(ctx, lookupJID); err == nil && !lidJID.IsEmpty() {
				info.LID = lidJID.ToNonAD().String()
				if contactErr != nil && store.Contacts != nil {
					if lidContact, err := store.Contacts.GetContact(ctx, lidJID); err == nil {
						applyVoIPContactInfo(&info, lidContact)
					}
				}
			}
		}
	}

	if phone, err := whatsapp.GetPhoneIfValid(info.Phone); err == nil {
		info.PhoneE164 = phone
	}
	if info.Title == "" {
		info.Title = firstVoIPText(info.FullName, info.BusinessName, info.PushName, info.FirstName)
	}
	return info
}

func applyVoIPContactInfo(target *voip.CallerInfo, source types.ContactInfo) {
	if target == nil {
		return
	}
	target.FullName = source.FullName
	target.BusinessName = source.BusinessName
	target.PushName = source.PushName
	target.FirstName = source.FirstName
	target.Title = ExtractContactName(source)
}

func firstVoIPText(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
