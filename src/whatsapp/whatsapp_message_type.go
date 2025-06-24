package whatsapp

import "encoding/json"

type WhatsappMessageType uint

const (
	UnknownMessageType WhatsappMessageType = iota
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

	// Messages that isn't important for this whatsapp service
	DiscardMessageType
)

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
	case DiscardMessageType:
		return "discard"
	}

	return "unknown"
}

func (s WhatsappMessageType) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
