package api

import (
	models "github.com/nocodeleaks/quepasa/models"
	runtime "github.com/nocodeleaks/quepasa/runtime"
)

func listPersistedServerRecords() ([]*models.QpServer, error) {
	return runtime.ListPersistedSessionRecords()
}

func findPersistedServerRecord(token string) (*models.QpServer, error) {
	return runtime.FindPersistedSessionRecord(token)
}
