package whatsapp

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
	}

	return "unknown"
}
