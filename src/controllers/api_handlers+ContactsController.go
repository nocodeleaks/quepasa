package controllers

import (
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
)

//region CONTROLLER - CONTACTS

func ContactsController(w http.ResponseWriter, r *http.Request) {

	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpContactsResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	contacts, err := server.GetContacts()
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.Total = len(contacts)
	response.Contacts = contacts
	RespondSuccess(w, response)
}

//endregion
