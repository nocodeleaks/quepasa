package api

import (
	"net/http"
	"sort"

	models "github.com/nocodeleaks/quepasa/models"
)

//region CONTROLLER - CONTACTS

// ContactsController retrieves all contacts from WhatsApp
//
//	@Summary		Get contacts
//	@Description	Retrieves a list of all WhatsApp contacts
//	@Tags			Contacts
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	models.QpContactsResponse
//	@Failure		400	{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/contacts [get]
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

	// Get contacts (works with active connection or cached data automatically)
	contacts, err := server.GetContacts()
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Sort contacts by ID to ensure consistent ordering
	sort.Slice(contacts, func(i, j int) bool {
		return contacts[i].Id < contacts[j].Id
	})

	response.Total = len(contacts)
	response.Contacts = contacts
	RespondSuccess(w, response)
}

//endregion
