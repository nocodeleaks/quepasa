package controllers

import (
	"fmt"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
)

/*
<summary>

	Find a whatsapp server by token passed on Url Path parameters

</summary>
*/
func GetServer(r *http.Request) (server *models.QpWhatsappServer, err error) {
	token := GetToken(r)
	return models.GetServerFromToken(token)
}

// <summary>Find a whatsapp server by token passed on Url Path parameters</summary>
func GetServerRespondOnError(w http.ResponseWriter, r *http.Request) (server *models.QpWhatsappServer, err error) {
	token := GetToken(r)
	server, err = models.GetServerFromToken(token)
	if err != nil {
		RespondNoContent(w, fmt.Errorf("token '%s' not found", token))
	}
	return
}
