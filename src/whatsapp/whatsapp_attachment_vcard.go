package whatsapp

func GenerateVCardAttachment(content []byte, filename string) (attach *WhatsappAttachment) {
	length := uint64(len(content))

	attach = &WhatsappAttachment{
		Mimetype:   "text/x-vcard",
		FileName:   filename,
		FileLength: length,
	}

	attach.SetContent(&content)
	return
}
