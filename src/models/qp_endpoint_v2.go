package models

// Obsolete, keep for compatibility with zammad
type QPEndpointV2 struct {
	ID        string `json:"id"`
	UserName  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Title     string `json:"title,omitempty"`
	Phone     string `json:"phone,omitempty"`
}
