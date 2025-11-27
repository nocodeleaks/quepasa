package api

import (
	"fmt"
	"net/http"
	"time"

	api "github.com/nocodeleaks/quepasa/api/models"
	models "github.com/nocodeleaks/quepasa/models"
)

//region CONTROLLER - HEALTH

// BasicHealthController - Simple health check without authentication
// Returns 200 OK if the application is running
//
//	@Summary		Health check
//	@Description	Basic health check endpoint to verify if the application is running
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	api.HealthResponse
//	@Security		ApiKeyAuth
//	@Router			/health [get]
func BasicHealthController(w http.ResponseWriter, r *http.Request) {
	response := &api.HealthResponse{
		QpResponse: models.QpResponse{
			Success: true,
			Status:  "application is running",
		},
		Timestamp: time.Now(),
		Version:   models.QpVersion,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	RespondInterface(w, response)
}

// HealthController provides health check with optional authentication
//
//	@Summary		Health check with optional authentication
//	@Description	Provides basic health check without auth, or detailed server info with token/master key
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	api.HealthResponse
//	@Router			/health [get]
func HealthController(w http.ResponseWriter, r *http.Request) {

	response := &api.HealthResponse{
		Timestamp: time.Now(),
		Version:   models.QpVersion,
	}

	if models.WhatsappService == nil {
		response.Success = false
		response.Status = "whatsapp service not initialized"
		RespondInterface(w, response)
		return
	}

	// Check if any authentication is provided
	hasAuth := false
	token := GetToken(r)
	master := IsMatchForMaster(r)

	if token != "" || master {
		hasAuth = true
	}

	// If no authentication provided, return basic health check
	if !hasAuth {
		response.Success = true
		response.Status = "application is running"
		RespondInterface(w, response)
		return
	}

	// Handle master key authentication first (higher priority)
	if master {
		healthItems := models.WhatsappService.GetHealth()
		response.Items = healthItems

		// Calculate statistics
		stats := calculateHealthStats(healthItems)
		response.Stats = &stats

		// Set success to true only if all servers are healthy
		response.Success = stats.Unhealthy == 0 && stats.Total > 0

		RespondInterface(w, response)
		return
	}

	// Handle token authentication (single server view)
	if token != "" {
		server, err := GetServer(r)
		if err != nil {
			// Return error for invalid token
			response.Success = false
			response.Status = fmt.Sprintf("invalid token: %s", err.Error())
			RespondInterface(w, response)
			return
		}

		// Return server-specific information
		state := server.GetState()
		response.Success = state.IsHealthy()
		response.Status = fmt.Sprintf("server state is %s", state.String())

		response.State = state
		response.StateCode = state.EnumIndex()
		response.Wid = server.Wid

		RespondInterface(w, response)
		return
	}

	// Fallback: basic health check if authentication is invalid
	response.Success = true
	response.Status = "application is running"
	RespondInterface(w, response)
}

// calculateHealthStats calculates statistics for all servers
func calculateHealthStats(items []models.QpHealthResponseItem) api.HealthStats {
	stats := api.HealthStats{
		Total: len(items),
	}

	for _, item := range items {
		if item.GetHealth() {
			stats.Healthy++
		} else {
			stats.Unhealthy++
		}
	}

	if stats.Total > 0 {
		stats.Percentage = float64(stats.Healthy) / float64(stats.Total) * 100
	}

	return stats
}

//endregion

func All[T any](ts []T, pred func(T) bool) bool {
	for _, t := range ts {
		if !pred(t) {
			return false
		}
	}
	return true
}
