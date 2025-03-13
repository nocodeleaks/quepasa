package controllers

import (
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
	"go.mau.fi/whatsmeow/types"
)

type QpGroupsResponse struct {
	models.QpResponse
	Total  int                `json:"total,omitempty"`
	Groups []*types.GroupInfo `json:"groups,omitempty"`
}

//region CONTROLLER - GET GROUP

func GetGroupController(w http.ResponseWriter, r *http.Request) {

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

//region CONTROLLER - FETCH ALL GROUPS

func FetchAllGroupsController(w http.ResponseWriter, r *http.Request) {
	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &QpGroupsResponse{}

	// Get the server from the request
	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Get all joined groups
	groups, err := server.GetJoinedGroups()
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.Total = len(groups)
	response.Groups = groups

	RespondSuccess(w, response)
}

//endregion

//region CONTROLLER - CREATE GROUP

func CreateGroupController(w http.ResponseWriter, r *http.Request) {

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
