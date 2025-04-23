package models

type LinkPreviewRequest struct {
	ChatId         string `json:"chat_id"`                      // Required: Chat to send to
	Text           string `json:"text"`                         // Required: Message text (must contain URL)
	TrackId        string `json:"trackid,omitempty"`            // Optional: For tracking the message
	FetchPreview   bool   `json:"fetch_preview,omitempty"`      // Optional: Whether to fetch preview (default true)
	CustomTitle    string `json:"custom_title,omitempty"`       // Optional: Override fetched title
	CustomDesc     string `json:"custom_description,omitempty"` // Optional: Override fetched description
	CustomThumbUrl string `json:"custom_thumb_url,omitempty"`   // Optional: Override fetched thumbnail
}
