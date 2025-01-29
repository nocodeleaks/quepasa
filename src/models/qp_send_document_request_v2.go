package models

type QPSendDocumentRequestV2 struct {
	Recipient  string         `json:"recipient,omitempty"`
	Message    string         `json:"message,omitempty"`
	Attachment QPAttachmentV1 `json:"attachment,omitempty"`
}

func (source *QPSendDocumentRequestV2) ToQpSendRequest() *QpSendRequest {
	request := &QpSendAnyRequest{}
	request.ChatId = source.Recipient
	request.Text = source.Message

	request.Mimetype = source.Attachment.MIME
	request.FileName = source.Attachment.FileName

	if len(source.Attachment.Base64) > 0 {
		request.Content = source.Attachment.Base64
		request.GenerateEmbedContent()

	} else if len(source.Attachment.Url) > 0 {
		request.Url = source.Attachment.Url
		request.GenerateUrlContent()
	}

	return &request.QpSendRequest
}
