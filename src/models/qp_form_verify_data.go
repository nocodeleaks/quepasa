package models

type QPFormVerifyData struct {
	PageTitle    string
	ErrorMessage string
	Bot          QPBot
	Protocol     string
	Host         string
	Destination  string
}
