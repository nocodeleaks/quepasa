package environment

// Metrics environment variable names
const (
	ENV_METRICS        = "METRICS"        // metrics endpoint enable/disable
	ENV_METRICS_PREFIX = "METRICS_PREFIX" // metrics endpoint path prefix (default: "metrics")
)

// MetricsSettings holds all Metrics configuration loaded from environment
type MetricsSettings struct {
	Enabled bool   `json:"enabled"`
	Prefix  string `json:"prefix"`
}

// NewMetricsSettings creates a new Metrics settings by loading all values from environment
func NewMetricsSettings() MetricsSettings {
	return MetricsSettings{
		Enabled: getEnvOrDefaultBool(ENV_METRICS, true),
		Prefix:  getEnvOrDefaultString(ENV_METRICS_PREFIX, "metrics"),
	}
}