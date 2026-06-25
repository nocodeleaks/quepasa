package whatsapp

type WhatsappMessageStatus string

const (
	WhatsappMessageStatusUnknown   WhatsappMessageStatus = ""
	WhatsappMessageStatusError     WhatsappMessageStatus = "error"
	WhatsappMessageStatusImported  WhatsappMessageStatus = "imported"
	WhatsappMessageStatusDelivered WhatsappMessageStatus = "delivered"
	WhatsappMessageStatusRead      WhatsappMessageStatus = "read"
)

func (source WhatsappMessageStatus) Uint32() uint {
	switch source {
	case WhatsappMessageStatusError:
		return 1
	case WhatsappMessageStatusImported:
		return 2
	case WhatsappMessageStatusDelivered:
		return 3
	case WhatsappMessageStatusRead:
		return 4
	}
	return 0
}
