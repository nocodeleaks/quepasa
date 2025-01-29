package models

type QPFormSendData struct {
	PageTitle    string
	MessageId    string
	ErrorMessage string
	Server       *QpServer
}
