package models

type QpIsOnWhatsappResponse struct {
	QpResponse
	Total      int      `json:"total"`
	Registered []string `json:"registered,omitempty"`
}
