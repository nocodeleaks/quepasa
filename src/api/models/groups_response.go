package api

import models "github.com/nocodeleaks/quepasa/models"

// GroupsResponse is the API transport shape for group collection endpoints.
type GroupsResponse struct {
	models.QpResponse
	Total  int           `json:"total,omitempty"`
	Groups []interface{} `json:"groups,omitempty"`
}

// SingleGroupResponse is the API transport shape for single-group reads.
type SingleGroupResponse struct {
	models.QpResponse
	Total     int         `json:"total,omitempty"`
	GroupInfo interface{} `json:"groupinfo,omitempty"`
}

// ParticipantResponse is the API transport shape for participant list mutations.
type ParticipantResponse struct {
	models.QpResponse
	Total        int           `json:"total,omitempty"`
	Participants []interface{} `json:"participants,omitempty"`
}

// RequestResponse is the API transport shape for join-request list operations.
type RequestResponse struct {
	models.QpResponse
	Total    int           `json:"total,omitempty"`
	Requests []interface{} `json:"requests,omitempty"`
}
