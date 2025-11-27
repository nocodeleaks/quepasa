package api

import (
	environment "github.com/nocodeleaks/quepasa/environment"
	models "github.com/nocodeleaks/quepasa/models"
)

// EnvironmentResponse represents environment settings response
type EnvironmentResponse struct {
	models.QpResponse
	Settings *environment.EnvironmentSettings        `json:"settings,omitempty"`
	Preview  *environment.EnvironmentSettingsPreview `json:"preview,omitempty"`
}
