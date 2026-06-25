package whatsmeow

import types "go.mau.fi/whatsmeow/types"

// UserInfoResponse represents the structured response for user information
type WhatsmeowUserInfoResponse struct {
	JID          string              `json:"jid"`
	LID          string              `json:"lid,omitempty"`
	Phone        string              `json:"phone,omitempty"`
	PhoneE164    string              `json:"phoneE164,omitempty"`
	Status       string              `json:"status,omitempty"`
	PictureID    string              `json:"pictureId,omitempty"`
	Devices      []types.JID         `json:"devices,omitempty"`
	VerifiedName *types.VerifiedName `json:"verifiedName,omitempty"`
	DisplayName  string              `json:"displayName,omitempty"`
	FullName     string              `json:"fullName,omitempty"`
	BusinessName string              `json:"businessName,omitempty"`
	PushName     string              `json:"pushName,omitempty"`
}
