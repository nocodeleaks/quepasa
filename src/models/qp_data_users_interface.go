package models

type QpDataUsersInterface interface {
	Count() (int, error)
	FindAll() ([]*QpUser, error)
	Find(string) (*QpUser, error)
	Exists(string) (bool, error)
	Check(string, string) (*QpUser, error)
	Create(username string, password string) (*QpUser, error)
	UpdatePassword(username string, password string) error
	UpdateUI(username string, ui string) error
	Delete(username string) error

	// FindByAPIKey resolves a user from the SHA-256 hash of their personal API key.
	FindByAPIKey(apikeyHash string) (*QpUser, error)
	// SetAPIKey stores (or rotates) the user's API key hash and stamps the rotation time.
	SetAPIKey(username string, apikeyHash string) error
	// ClearAPIKey revokes the user's API key.
	ClearAPIKey(username string) error
}
