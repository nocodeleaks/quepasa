package models

type QpGroupsResponse struct {
	QpResponse
	Total  int           `json:"total,omitempty"`
	Groups []interface{} `json:"groups,omitempty"`
}

type QpSingleGroupResponse struct {
	QpResponse
	Total     int         `json:"total,omitempty"`
	GroupInfo interface{} `json:"groupinfo,omitempty"`
}

type QpParticipantResponse struct {
	QpResponse
	Total        int           `json:"total,omitempty"`
	Participants []interface{} `json:"participants,omitempty"`
}

type QpRequestResponse struct {
	QpResponse
	Total    int           `json:"total,omitempty"`
	Requests []interface{} `json:"requests,omitempty"`
}
