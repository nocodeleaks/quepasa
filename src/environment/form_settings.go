package environment

// Form environment variable names
const (
	ENV_FORM              = "FORM"              // enable/disable form interface
	ENV_FORM_PREFIX       = "FORM_PREFIX"       // form endpoint path prefix (default: "form")
	ENV_SERVERS_VIEW_MODE = "SERVERS_VIEW_MODE" // servers list view mode: "card" or "table" (default: "card")
)

// FormSettings holds all Form configuration loaded from environment
type FormSettings struct {
	Enabled         bool   `json:"enabled"`
	Prefix          string `json:"prefix"`
	ServersViewMode string `json:"serversViewMode"` // "card" or "table"
}

// NewFormSettings creates a new Form settings by loading all values from environment
func NewFormSettings() FormSettings {
	return FormSettings{
		Enabled:         getEnvOrDefaultBool(ENV_FORM, true),
		Prefix:          getEnvOrDefaultString(ENV_FORM_PREFIX, "form"),
		ServersViewMode: getEnvOrDefaultString(ENV_SERVERS_VIEW_MODE, "card"),
	}
}
