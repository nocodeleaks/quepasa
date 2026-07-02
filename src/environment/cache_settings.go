package environment

const (
	ENV_CACHE_BACKEND       = "CACHE_BACKEND"
	ENV_CACHE_DISK_PATH     = "CACHE_DISK_PATH"
	ENV_CACHE_INIT_FALLBACK = "CACHE_INIT_FALLBACK"
)

type CacheSettings struct {
	Backend       string `json:"backend"`
	DiskPath      string `json:"disk_path"`
	InitFallback  bool   `json:"init_fallback"`
	HotWindowDays uint32 `json:"hot_window_days"`
}

func NewCacheSettings() CacheSettings {
	s := CacheSettings{
		Backend:       getEnvOrDefaultString(ENV_CACHE_BACKEND, "memory"),
		DiskPath:      getEnvOrDefaultString(ENV_CACHE_DISK_PATH, ""),
		InitFallback:  getEnvOrDefaultBool(ENV_CACHE_INIT_FALLBACK, true),
		HotWindowDays: getEnvOrDefaultUint32(ENV_CACHEDAYS, 90),
	}
	// hot-window: never 0 (0 = unbounded redis TTL = RAM leak). Default 90d.
	if s.HotWindowDays == 0 {
		s.HotWindowDays = 90
	}
	return s
}
