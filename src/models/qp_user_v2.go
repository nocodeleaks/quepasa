package models

// Usuário no formato QuePasa (Telegram)
// Not in use for now
type QpUserV2 struct {
	ID                      string `json:"id"`
	IsBot                   bool   `json:"is_bot,omitempty"`
	FirstName               string `json:"first_name,omitempty"`
	LastName                string `json:"last_name,omitempty"`
	UserName                string `json:"username,omitempty"`
	LanguageCode            string `json:"language_code,omitempty"`
	CanJoinGroups           bool   `json:"can_join_groups,omitempty"`
	CanReadAllGroupMessages bool   `json:"can_read_all_group_messages,omitempty"`
	SupportsInlineQueries   bool   `json:"supports_inline_queries,omitempty"`
}
