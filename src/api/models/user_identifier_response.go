package api

import (
	models "github.com/nocodeleaks/quepasa/models"
)

type UserIdentifierResponse struct {
	models.QpResponse
	UserIdentifierRequest
}
