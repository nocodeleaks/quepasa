package api

import (
	metrics "github.com/nocodeleaks/quepasa/metrics"
)

// APIProcessingTime is the public API processing time metric
var APIProcessingTime = metrics.CreateHistogramVecRecorder(
	"quepasa_api_request_duration_seconds",
	"Time spent processing API requests",
	[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
	[]string{"method", "endpoint", "status_code"},
)

// ObserveAPIRequestDuration records API request processing time
func ObserveAPIRequestDuration(method, endpoint, statusCode string, duration float64) {
	if APIProcessingTime != nil {
		histogram := APIProcessingTime.WithLabelValues(method, endpoint, statusCode)
		histogram.Observe(duration)
	}
}
