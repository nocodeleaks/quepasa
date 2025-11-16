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

// HealthController provides detailed health check with authentication support
//
//	@Summary		Detailed health check
//	@Description	Provides detailed health information for WhatsApp servers with authentication support
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	api.HealthResponse
//	@Router			/healthapi [get]
func HealthController(w http.ResponseWriter, r *http.Request) {

	response := &api.HealthResponse{
		Timestamp: time.Now(),
		Version:   models.QpVersion,
	}

	master := IsMatchForMaster(r)
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
	} else {
		// Check for user authentication headers first
		username := r.Header.Get("X-QUEPASA-USER")
		password := r.Header.Get("X-QUEPASA-PASSWORD")

		if username != "" && password != "" {
			// Authenticate user
			user, err := models.WhatsappService.GetUser(username, password)
			if err != nil {
				response.ParseError(fmt.Errorf("authentication failed: %s", err.Error()))
				RespondInterface(w, response)
				return
			}

			// Get all servers for this user
			userServers := models.WhatsappService.GetServersForUser(user.Username)

			var healthItems []models.QpHealthResponseItem
			for _, server := range userServers {
				item := models.ToHealthReponseItem(server)
				healthItems = append(healthItems, item)
			}

			response.Items = healthItems

			// Calculate statistics for user servers
			stats := calculateHealthStats(healthItems)
			response.Stats = &stats

			// Set success to true only if all user servers are healthy
			response.Success = stats.Unhealthy == 0 && stats.Total > 0

			RespondInterface(w, response)
			return
		}

		// Fallback to single server authentication by token
		server, err := GetServer(r)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		state := server.GetStatus()
		response.Success = state.IsHealthy()
		response.Status = fmt.Sprintf("server state is %s", state.String())
		response.State = state
		response.StateCode = state.EnumIndex()

		RespondInterface(w, response)
		return
	}
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

// boolToInt converts boolean to int (true = 1, false = 0)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// boolToFloat converts boolean to float64 (true = 1.0, false = 0.0)
func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
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
