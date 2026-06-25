package whatsapp

type IWhatsappChatId interface {

	// E164 Phone without trailing + or GroupID with -
	// Ex: 5521967609494
	// Ex: 5521967609494-1445779956
	GetChatId() string
}
