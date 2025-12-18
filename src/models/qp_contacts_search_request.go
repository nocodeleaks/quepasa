package models

// QpContactsSearchRequest represents the request body for contact search
type QpContactsSearchRequest struct {
	Query    string `json:"query,omitempty"`     // Search in title and phone (optional)
	HasTitle *bool  `json:"has_title,omitempty"` // Filter contacts with/without title: true=with title, false=without title, null=no filter
	HasLid   *bool  `json:"has_lid,omitempty"`   // Filter contacts with/without LID: true=with LID, false=without LID, null=no filter
	Phone    string `json:"phone,omitempty"`     // Search by specific phone (optional)
}
