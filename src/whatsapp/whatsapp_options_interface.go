package whatsapp

type IWhatsappOptions interface {
	GetOptions() *WhatsappOptions
	Save(reason string) error
}
