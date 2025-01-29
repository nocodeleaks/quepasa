package models

type QpInfoResponseV2 struct {
	QpResponse
	Id        string `json:"id"`
	Number    string `json:"number"`
	Username  string `json:"username" validate:"max=255"`
	FirstName string `json:"first_name" validate:"max=255"`
	LastName  string `json:"last_name" validate:"max=255"`
}
