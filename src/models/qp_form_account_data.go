package models

type QPFormAccountData struct {
	PageTitle    string
	ErrorMessage string
	Version      string
	Servers      map[string]*QpWhatsappServer
	User         QpUser
}
