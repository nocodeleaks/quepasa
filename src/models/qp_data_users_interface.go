package models

type QpDataUsersInterface interface {
	Count() (int, error)
	FindAll() ([]*QpUser, error)
	Find(string) (*QpUser, error)
	Exists(string) (bool, error)
	Check(string, string) (*QpUser, error)
	Create(username string, password string) (*QpUser, error)
	UpdatePassword(username string, password string) error
	Delete(username string) error
}
