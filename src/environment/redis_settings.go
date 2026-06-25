package environment

const (
	ENV_REDIS_HOST                  = "REDIS_HOST"
	ENV_REDIS_PORT                  = "REDIS_PORT"
	ENV_REDIS_USERNAME              = "REDIS_USERNAME"
	ENV_REDIS_PASSWORD              = "REDIS_PASSWORD"
	ENV_REDIS_DATABASE              = "REDIS_DATABASE"
	ENV_REDIS_KEY_PREFIX            = "REDIS_KEY_PREFIX"
	ENV_REDIS_POOL_SIZE             = "REDIS_POOL_SIZE"
	ENV_REDIS_MAX_RETRIES           = "REDIS_MAX_RETRIES"
	ENV_REDIS_DIAL_TIMEOUT_SECONDS  = "REDIS_DIAL_TIMEOUT_SECONDS"
	ENV_REDIS_READ_TIMEOUT_SECONDS  = "REDIS_READ_TIMEOUT_SECONDS"
	ENV_REDIS_WRITE_TIMEOUT_SECONDS = "REDIS_WRITE_TIMEOUT_SECONDS"
)

type RedisSettings struct {
	Host                string `json:"host"`
	Port                uint32 `json:"port"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	Database            uint32 `json:"database"`
	KeyPrefix           string `json:"key_prefix"`
	PoolSize            uint32 `json:"pool_size"`
	MaxRetries          uint32 `json:"max_retries"`
	DialTimeoutSeconds  uint32 `json:"dial_timeout_seconds"`
	ReadTimeoutSeconds  uint32 `json:"read_timeout_seconds"`
	WriteTimeoutSeconds uint32 `json:"write_timeout_seconds"`
}

func NewRedisSettings() RedisSettings {
	return RedisSettings{
		Host:                getEnvOrDefaultString(ENV_REDIS_HOST, ""),
		Port:                getEnvOrDefaultUint32(ENV_REDIS_PORT, 6379),
		Username:            getEnvOrDefaultString(ENV_REDIS_USERNAME, ""),
		Password:            getEnvOrDefaultString(ENV_REDIS_PASSWORD, ""),
		Database:            getEnvOrDefaultUint32(ENV_REDIS_DATABASE, 0),
		KeyPrefix:           getEnvOrDefaultString(ENV_REDIS_KEY_PREFIX, "quepasa"),
		PoolSize:            getEnvOrDefaultUint32(ENV_REDIS_POOL_SIZE, 10),
		MaxRetries:          getEnvOrDefaultUint32(ENV_REDIS_MAX_RETRIES, 3),
		DialTimeoutSeconds:  getEnvOrDefaultUint32(ENV_REDIS_DIAL_TIMEOUT_SECONDS, 5),
		ReadTimeoutSeconds:  getEnvOrDefaultUint32(ENV_REDIS_READ_TIMEOUT_SECONDS, 3),
		WriteTimeoutSeconds: getEnvOrDefaultUint32(ENV_REDIS_WRITE_TIMEOUT_SECONDS, 3),
	}
}

func (settings RedisSettings) Enabled() bool {
	return len(settings.Host) > 0
}
