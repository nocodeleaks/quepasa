package environment

// Swagger environment variable names
const (
	ENV_SWAGGER        = "SWAGGER"        // swagger UI enable/disable
	ENV_SWAGGER_PREFIX = "SWAGGER_PREFIX" // swagger UI path prefix (default: "swagger")
)

// SwaggerSettings holds all Swagger configuration loaded from environment
type SwaggerSettings struct {
	Enabled bool   `json:"enabled"`
	Prefix  string `json:"prefix"`
}

// NewSwaggerSettings creates a new Swagger settings by loading all values from environment
func NewSwaggerSettings() SwaggerSettings {
	return SwaggerSettings{
		Enabled: getEnvOrDefaultBool(ENV_SWAGGER, true),
		Prefix:  getEnvOrDefaultString(ENV_SWAGGER_PREFIX, "swagger"),
	}
}
