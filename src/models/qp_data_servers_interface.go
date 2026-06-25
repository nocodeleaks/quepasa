package models

type QpDataServersInterface interface {
	FindAll() []*QpServer
	FindByToken(string) (*QpServer, error)
	FindForUser(string, string) (*QpServer, error)
	Exists(string) (bool, error)

	Add(*QpServer) error
	Update(*QpServer) error
	Delete(string) error
}
