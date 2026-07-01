package whatsapp

import "encoding/json"

type WhatsappMessageType uint

const (
	// Messages that isn't important for this whatsapp service
	// These are usually used for internal purposes or debugging
	// and should not be processed further
	// It must contains a reason for the discard on Debug property
	UnhandledMessageType WhatsappMessageType = iota
	ViewOnceMessageType
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
	StickerMessageType
)

func (s WhatsappMessageType) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON reverses MarshalJSON so JSON-backed stores (redis/disk) can read
// the string form back. Also tolerates numeric encodings for backward compat.
func (t *WhatsappMessageType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		var n uint
		if nerr := json.Unmarshal(data, &n); nerr == nil {
			*t = WhatsappMessageType(n)
			return nil
		}
		return err
	}
	switch s {
	case "image":
		*t = ImageMessageType
	case "document":
		*t = DocumentMessageType
	case "audio":
		*t = AudioMessageType
	case "video":
		*t = VideoMessageType
	case "text":
		*t = TextMessageType
	case "location":
		*t = LocationMessageType
	case "contact":
		*t = ContactMessageType
	case "call":
		*t = CallMessageType
	case "system":
		*t = SystemMessageType
	case "group":
		*t = GroupMessageType
	case "revoke":
		*t = RevokeMessageType
	case "poll":
		*t = PollMessageType
	case "sticker":
		*t = StickerMessageType
	case "view_once":
		*t = ViewOnceMessageType
	default:
		*t = UnhandledMessageType
	}
	return nil
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
	case StickerMessageType:
		return "sticker"
	case ViewOnceMessageType:
		return "view_once"
	}

	// If the type is not recognized, return "unhandled"
	return "unhandled"
}
