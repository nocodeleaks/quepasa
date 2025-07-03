package whatsapp

import "encoding/json"

type WhatsappMessageType uint

const (
	// Messages that isn't important for this whatsapp service
	// These are usually used for internal purposes or debugging
	// and should not be processed further
	// It must contains a reason for the discard on Debug property
	UnhandledMessageType WhatsappMessageType = iota
	ImageMessageType
	DocumentMessageType
	AudioMessageType
	VideoMessageType
	TextMessageType
	LocationMessageType
	ContactMessageType
	CallMessageType
	SystemMessageType
	GroupMessageType
	RevokeMessageType
	PollMessageType
)

func (s WhatsappMessageType) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (Type WhatsappMessageType) String() string {
	switch Type {
	case ImageMessageType:
		return "image"
	case DocumentMessageType:
		return "document"
	case AudioMessageType:
		return "audio"
	case VideoMessageType:
		return "video"
	case TextMessageType:
		return "text"
	case LocationMessageType:
		return "location"
	case ContactMessageType:
		return "contact"
	case CallMessageType:
		return "call"
	case SystemMessageType:
		return "system"
	case GroupMessageType:
		return "group"
	case RevokeMessageType:
		return "revoke"
	case PollMessageType:
		return "poll"
	}

	// If the type is not recognized, return "unhandled"
	return "unhandled"
}
