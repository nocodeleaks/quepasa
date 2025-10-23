package metrics

import (
	"github.com/go-chi/chi/v5"
	environment "github.com/nocodeleaks/quepasa/environment"
	webserver "github.com/nocodeleaks/quepasa/webserver"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func init() {
	// Automatically registers the Swagger configuration in the webserver
	// This allows Swagger to be configured without the webserver module
	// needing to know specifically about Swagger
	webserver.RegisterRouterConfigurator(Configure)
}

// MetricsEnabled is a global flag to enable/disable metrics
// This is automatically set based on environment configuration
var MetricsEnabled = environment.Settings.Metrics.Enabled

// ConfigureMetrics configures the metrics endpoint in the provided router
// This function can be called from webserver to set up metrics without
// creating circular dependencies
func Configure(r chi.Router) {
	if MetricsEnabled {
		ServeMetrics(r)

		if environment.Settings.Metrics.Dashboard.Enabled {
			ServeDashboard(r)
		}
	}
}

// ServeMetrics serves the Prometheus metrics endpoint
func ServeMetrics(r chi.Router) {
	prefix := environment.Settings.Metrics.Prefix
	r.Handle("/"+prefix, promhttp.Handler())
}

func ServeDashboard(r chi.Router) {
	log.Debug("starting dashboard service")
	prefix := environment.Settings.Metrics.Dashboard.Prefix
	r.Get("/"+prefix, DashboardHandler)
}
