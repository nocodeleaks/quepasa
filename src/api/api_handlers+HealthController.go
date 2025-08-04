package api

import (
	"fmt"
	"net/http"
	"time"

	api "github.com/nocodeleaks/quepasa/api/models"
	models "github.com/nocodeleaks/quepasa/models"
)

//region CONTROLLER - HEALTH

func HealthController(w http.ResponseWriter, r *http.Request) {

	response := &api.HealthResponse{Timestamp: time.Now()}

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

		status := server.GetStatus()
		response.Success = status.IsHealthy()
		response.Status = fmt.Sprintf("server status is %s", status.String())

		// Single server stats
		response.Stats = &api.HealthStats{
			Total:      1,
			Healthy:    boolToInt(status.IsHealthy()),
			Unhealthy:  boolToInt(!status.IsHealthy()),
			Percentage: boolToFloat(status.IsHealthy()) * 100,
		}

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
