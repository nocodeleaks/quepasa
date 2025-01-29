package whatsapp

type IWhatsappMessage interface {
	IWhatsappChatId

	GetId() string

	// Clear text message or html encoded
	GetText() string

	// Check if that msg has a valid attachment
	HasAttachment() bool

	// Get if exists bytes of attachments
	GetAttachment() *WhatsappAttachment

	// Original message from source service
	GetSource() interface{}

	GetParticipantId() string
}
