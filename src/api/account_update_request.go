package api

// AccountUpdateRequest represents the request body for account updates.
type AccountUpdateRequest struct {
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}