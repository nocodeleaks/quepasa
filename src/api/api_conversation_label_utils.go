package api

import (
	models "github.com/nocodeleaks/quepasa/models"
	runtime "github.com/nocodeleaks/quepasa/runtime"
)

func findConversationLabelStore() (models.QpDataConversationLabelsInterface, error) {
	return runtime.GetConversationLabelStore()
}
