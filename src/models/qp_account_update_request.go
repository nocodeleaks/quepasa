package models

type QpAccountUpdateRequest struct {
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}
