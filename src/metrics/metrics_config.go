package metrics

import (
	"github.com/go-chi/chi/v5"
	environment "github.com/nocodeleaks/quepasa/environment"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ConfigureMetrics configures the metrics endpoint in the provided router
// This function can be called from webserver to set up metrics without
// creating circular dependencies
func ConfigureMetrics(r chi.Router) {
	if environment.Settings.Metrics.Enabled {
		ServeMetrics(r)
	}
}

// ServeMetrics serves the Prometheus metrics endpoint
func ServeMetrics(r chi.Router) {
	prefix := environment.Settings.Metrics.Prefix
	r.Handle("/"+prefix, promhttp.Handler())
}
