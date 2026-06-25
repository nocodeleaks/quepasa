package models

// Method for control messages
type IQpMessages interface {

	// Delete|Revoke|Cancel message
	Revoke(string) error

	// Download message attachments
	Download(string) ([]byte, error)
}
