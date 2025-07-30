package api

type UserIdentifierRequest struct {
	Phone string `json:"phone,omitempty"`
	LId   string `json:"lid,omitempty"`
}
