package models

type QpServerV2 struct {
	ID              string `json:"id"`
	Verified        bool   `json:"is_verified"`
	Token           string `json:"token"`
	UserID          string `json:"user_id,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
	Devel           bool   `json:"devel"`
	Version         string `json:"version,omitempty"`
	HandleGroups    bool   `json:"handlegroups,omitempty"`
	HandleBroadcast bool   `json:"handlebroadcast,omitempty"`
}
