package models

type QpInfoResponse struct {
	QpResponse
	Server *QpWhatsappServer `json:"server,omitempty"`
}

func (source *QpInfoResponse) ParseSuccess(server *QpWhatsappServer) {
	source.QpResponse.ParseSuccess("follow server information")
	source.Server = server
}

func (source *QpInfoResponse) PatchSuccess(server *QpWhatsappServer, message string) {
	source.QpResponse.ParseSuccess(message)
	source.Server = server
}
