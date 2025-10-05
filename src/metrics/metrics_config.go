package metrics

import (
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ConfigureMetrics configures the metrics endpoint in the provided router
// This function can be called from webserver to set up metrics without
// creating circular dependencies
func ConfigureMetrics(r chi.Router) {
	r.Handle("/metrics", promhttp.Handler())
}