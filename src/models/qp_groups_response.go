package models

import types "go.mau.fi/whatsmeow/types"

type QpGroupsResponse struct {
	QpResponse
	Total  int                `json:"total,omitempty"`
	Groups []*types.GroupInfo `json:"groups,omitempty"`
}

type QpSingleGroupResponse struct {
	QpResponse
	Total     int                `json:"total,omitempty"`
	GroupInfo []*types.GroupInfo `json:"groupinfo,omitempty"`
}
