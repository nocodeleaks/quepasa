package api

import (
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
)

// GetSession returns the live WhatsApp session associated with the token in the request.
func GetSession(r *http.Request) (*models.QpWhatsappSession, error) {
	return models.GetSessionFromToken(GetToken(r))
}

// GetSessionRespondOnError mirrors GetServerRespondOnError while exposing session naming.
func GetSessionRespondOnError(w http.ResponseWriter, r *http.Request) (*models.QpWhatsappSession, error) {
	return GetServerRespondOnError(w, r)
}

// GetSessionFromMaster returns the first available session for a valid master-key request.
func GetSessionFromMaster(r *http.Request) (*models.QpWhatsappSession, error) {
	return GetServerFromMaster(r)
}
