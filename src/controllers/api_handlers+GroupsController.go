package controllers

import (
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
	types "go.mau.fi/whatsmeow/types"
)

//region CONTROLLER - GET GROUP

func GetGroupController(w http.ResponseWriter, r *http.Request) {

	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpSingleGroupResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	groupId := models.GetRequestParameter(r, "groupId")
	group, err := server.GetGroupInfo(groupId)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}
	response.GroupInfo = []*types.GroupInfo{group}

	RespondSuccess(w, response)
}

//endregion

//region CONTROLLER - FETCH ALL GROUPS

func FetchAllGroupsController(w http.ResponseWriter, r *http.Request) {
	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpGroupsResponse{}

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

	// Set the response with the groups and the amount of groups
	response.Total = len(groups)
	response.Groups = groups

	RespondSuccess(w, response)
}

//endregion

//region CONTROLLER - CREATE GROUP

func CreateGroupController(w http.ResponseWriter, r *http.Request) {

	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpGroupsResponse{}

	_, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	RespondSuccess(w, response)
}

//endregion
