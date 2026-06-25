package whatsapp

type WhatsappError interface {
	//Error() string
	Unauthorized() bool
}
