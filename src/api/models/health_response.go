package api

import (
	"time"

	models "github.com/nocodeleaks/quepasa/models"
)

type HealthResponse struct {
	models.QpResponse
	Items     []models.QpHealthResponseItem `json:"items,omitempty"`
	Stats     *HealthStats                  `json:"stats,omitempty"`
	Timestamp time.Time                     `json:"timestamp"`
}
