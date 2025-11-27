package api

import (
	"fmt"
	"net/http"
	"time"

	api "github.com/nocodeleaks/quepasa/api/models"
	library "github.com/nocodeleaks/quepasa/library"
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
	}

	RespondInterface(w, response)
}

// HealthController provides health check with optional authentication
//
//	@Summary		Health check with optional authentication
//	@Description	Provides basic health check without auth, detailed server info with token/master key, or user's servers with username/password
//	@Tags			Health
//	@Produce		json
//	@Param			X-QUEPASA-USER		header		string	false	"Username for user authentication"
//	@Param			X-QUEPASA-PASSWORD	header		string	false	"Password for user authentication"
//	@Success		200					{object}	api.HealthResponse
//	@Router			/health [get]
func HealthController(w http.ResponseWriter, r *http.Request) {

	response := &api.HealthResponse{
		Timestamp: time.Now(),
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
	username := r.Header.Get("X-QUEPASA-USER")
	password := r.Header.Get("X-QUEPASA-PASSWORD")

	if token != "" || master || (username != "" && password != "") {
		hasAuth = true
	}

	// If no authentication provided, return basic health check
	if !hasAuth {
		response.Success = true
		response.Status = "application is running"
		
		// Add application uptime
		uptime := library.Duration(time.Since(models.ApplicationStartTime))
		response.Uptime = &uptime
		
		RespondSuccess(w, response)
		return
	}

	// Handle master key authentication first (higher priority)
	if master {
		healthItems := models.WhatsappService.GetHealth()
		response.Items = healthItems

		// Calculate statistics
		stats := calculateHealthStats(healthItems)
		response.Stats = &stats

		// Add application uptime
		uptime := library.Duration(time.Since(models.ApplicationStartTime))
		response.Uptime = &uptime

		// Set success to true only if all servers are healthy
		response.Success = stats.Unhealthy == 0 && stats.Total > 0

		RespondInterface(w, response)
		return
	}

	// Handle user/password authentication (user's servers view)
	if username != "" && password != "" {
		// Validate user credentials
		user, err := models.WhatsappService.DB.Users.Check(username, password)
		if err != nil {
			response.Success = false
			response.Status = fmt.Sprintf("invalid credentials: %s", err.Error())
			RespondSuccess(w, response)
			return
		}

		// Get all servers for this user
		var userServers []models.QpHealthResponseItem
		for _, server := range models.WhatsappService.Servers {
			if server.User == user.Username {
				userServers = append(userServers, models.ToHealthReponseItem(server))
			}
		}

		response.Items = userServers

		// Calculate statistics for user's servers
		stats := calculateHealthStats(userServers)
		response.Stats = &stats

		// Add application uptime
		uptime := library.Duration(time.Since(models.ApplicationStartTime))
		response.Uptime = &uptime

		// Set success based on user's servers health
		response.Success = stats.Unhealthy == 0 && stats.Total > 0
		if stats.Total == 0 {
			response.Success = true
			response.Status = "no servers found for user"
		}

		RespondInterface(w, response)
		return
	}

	// Handle token authentication (single server view)
	if token != "" {
		server, err := GetServer(r)
		if err != nil {
			// Return error for invalid token - always 200 OK but Success=false
			response.Success = false
			response.Status = fmt.Sprintf("invalid token: %s", err.Error())
			RespondSuccess(w, response)
			return
		}

		// Return server-specific information - always 200 OK
		state := server.GetState()
		response.Success = state.IsHealthy()
		response.Status = fmt.Sprintf("server state is %s", state.String())

		response.State = &state
		response.Wid = server.Wid
		
		// Add server uptime
		if !server.Timestamps.Start.IsZero() {
			uptime := library.Duration(time.Since(server.Timestamps.Start))
			response.Uptime = &uptime
		}

		RespondSuccess(w, response)
		return
	}

	// Fallback: basic health check if authentication is invalid
	response.Success = true
	response.Status = "application is running"
	RespondSuccess(w, response)
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
