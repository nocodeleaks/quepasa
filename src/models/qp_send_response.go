package models

type QpSendResponse struct {
	QpResponse
	Message *QpSendResponseMessage `json:"message,omitempty"`
}

func (source *QpSendResponse) ParseSuccess(message *QpSendResponseMessage) {
	source.QpResponse.ParseSuccess("sended with success")
	source.Message = message
}
