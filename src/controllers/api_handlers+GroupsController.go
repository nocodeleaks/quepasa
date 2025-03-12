package controllers

import (
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
)

type QpGroupsResponse struct {
	models.QpResponse
}

//region CONTROLLER - CONTACTS

func GroupsController(w http.ResponseWriter, r *http.Request) {

	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &QpGroupsResponse{}

	_, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	RespondSuccess(w, response)
}

//endregion
