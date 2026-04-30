package api

import (
	models "github.com/nocodeleaks/quepasa/models"
	runtime "github.com/nocodeleaks/quepasa/runtime"
)

func findPersistedUser(username string) (*models.QpUser, error) {
	return runtime.FindPersistedUser(username)
}

func authenticatePersistedUser(username string, password string) (*models.QpUser, error) {
	return runtime.AuthenticateUser(username, password)
}

func updatePersistedUserPassword(username string, password string) error {
	return runtime.UpdatePersistedUserPassword(username, password)
}
