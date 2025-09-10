package api

import (
	"fmt"
	"net/http"
	"time"

	api "github.com/nocodeleaks/quepasa/api/models"
	models "github.com/nocodeleaks/quepasa/models"
	"github.com/prometheus/client_golang/prometheus"
)

//region CONTROLLER - HEALTH

// BasicHealthController - Simple health check without authentication
// Returns 200 OK if the application is running
func BasicHealthController(w http.ResponseWriter, r *http.Request) {
	response := &api.HealthResponse{
		QpResponse: models.QpResponse{
			Success: true,
			Status:  "application is running",
		},
		Timestamp: time.Now(),
		Queue:     getWebhookQueueStats(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	RespondInterface(w, response)
}

func HealthController(w http.ResponseWriter, r *http.Request) {

	response := &api.HealthResponse{Timestamp: time.Now()}

	master := IsMatchForMaster(r)
	if master {
		healthItems := models.WhatsappService.GetHealth()
		response.Items = healthItems

		// Calculate statistics
		stats := calculateHealthStats(healthItems)
		response.Stats = &stats

		// Add queue stats
		response.Queue = getWebhookQueueStats()

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
				response.Queue = getWebhookQueueStats()
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

			// Add queue stats
			response.Queue = getWebhookQueueStats()

			// Set success to true only if all user servers are healthy
			response.Success = stats.Unhealthy == 0 && stats.Total > 0

			RespondInterface(w, response)
			return
		}

		// Fallback to single server authentication by token
		server, err := GetServer(r)
		if err != nil {
			response.ParseError(err)
			response.Queue = getWebhookQueueStats()
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

		// Add queue stats
		response.Queue = getWebhookQueueStats()

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

// getWebhookQueueStats retrieves current webhook queue statistics
func getWebhookQueueStats() *api.WebhookQueueStats {
	if models.WebhookQueueClientInstance == nil {
		return &api.WebhookQueueStats{
			Enabled: false,
		}
	}

	queueStatus := models.WebhookQueueClientInstance.GetQueueStatus()

	// Get metric values from global registry
	processedTotal := getMetricValue("quepasa_webhook_queue_processed_total")
	discardedTotal := getMetricValue("quepasa_webhook_queue_discarded_total")
	retriesTotal := getMetricValue("quepasa_webhook_queue_retries_total")
	completedTotal := getMetricValue("quepasa_webhook_queue_completed_total")
	failedTotal := getMetricValue("quepasa_webhook_queue_failed_total")

	return &api.WebhookQueueStats{
		Enabled:         queueStatus["is_enabled"].(bool),
		CurrentSize:     queueStatus["current_size"].(int),
		MaxSize:         queueStatus["max_size"].(int),
		Utilization:     queueStatus["utilization"].(float64),
		ProcessingDelay: queueStatus["processing_delay"].(string),
		Workers:         queueStatus["workers"].(int),
		ProcessedTotal:  processedTotal,
		DiscardedTotal:  discardedTotal,
		RetriesTotal:    retriesTotal,
		CompletedTotal:  completedTotal,
		FailedTotal:     failedTotal,
	}
}

// getMetricValue retrieves the current value of a Prometheus metric by name
func getMetricValue(metricName string) float64 {
	// Use the default registry to gather metrics
	metrics, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return 0
	}

	for _, metricFamily := range metrics {
		if *metricFamily.Name == metricName {
			for _, metric := range metricFamily.Metric {
				if metric.Counter != nil {
					return *metric.Counter.Value
				}
				if metric.Gauge != nil {
					return *metric.Gauge.Value
				}
			}
		}
	}
	return 0
}
