package environment

const (
	ENV_CACHE_BACKEND       = "CACHE_BACKEND"
	ENV_CACHE_DISK_PATH     = "CACHE_DISK_PATH"
	ENV_CACHE_INIT_FALLBACK = "CACHE_INIT_FALLBACK"
)

type CacheSettings struct {
	Backend      string `json:"backend"`
	DiskPath     string `json:"disk_path"`
	InitFallback bool   `json:"init_fallback"`
}

func NewCacheSettings() CacheSettings {
	return CacheSettings{
		Backend:      getEnvOrDefaultString(ENV_CACHE_BACKEND, "memory"),
		DiskPath:     getEnvOrDefaultString(ENV_CACHE_DISK_PATH, ""),
		InitFallback: getEnvOrDefaultBool(ENV_CACHE_INIT_FALLBACK, true),
	}
}
