package api

// ContactSearchRequest represents the request body for contact search.
type ContactSearchRequest struct {
	Query    string `json:"query,omitempty"`
	HasTitle *bool  `json:"has_title,omitempty"`
	HasLid   *bool  `json:"has_lid,omitempty"`
	Phone    string `json:"phone,omitempty"`
}