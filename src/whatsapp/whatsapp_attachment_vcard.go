package whatsapp

func GenerateVCardAttachment(content []byte, filename string) (attach *WhatsappAttachment) {
	length := uint64(len(content))

	attach = &WhatsappAttachment{
		CanDownload: false,
		Mimetype:    "text/x-vcard",
		FileName:    filename,
		FileLength:  length,
	}

	attach.SetContent(&content)
	return
}
