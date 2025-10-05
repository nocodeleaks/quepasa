package environment

// Dashboard environment variable names
const (
	ENV_DASHBOARD        = "DASHBOARD"        // dashboard endpoint enable/disable
	ENV_DASHBOARD_PREFIX = "DASHBOARD_PREFIX" // dashboard endpoint path prefix (default: "dashboard")
)

// DashboardSettings holds all Dashboard configuration loaded from environment
type DashboardSettings struct {
	Enabled bool   `json:"enabled"`
	Prefix  string `json:"prefix"`
}

// NewDashboardSettings creates a new Dashboard settings by loading all values from environment
func NewDashboardSettings() DashboardSettings {
	return DashboardSettings{
		Enabled: getEnvOrDefaultBool(ENV_DASHBOARD, true),
		Prefix:  getEnvOrDefaultString(ENV_DASHBOARD_PREFIX, "dashboard"),
	}
}