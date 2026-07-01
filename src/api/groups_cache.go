package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	cache "github.com/nocodeleaks/quepasa/cache"
	environment "github.com/nocodeleaks/quepasa/environment"
	redis "github.com/redis/go-redis/v9"
)

// Persistent (redis) cache for the joined-groups list, so the Groups page does not
// hit WhatsApp live on every load. Falls back to live fetch when redis is not the
// configured cache backend or is unreachable.

var (
	groupsCacheOnce   sync.Once
	groupsCacheClient *redis.Client
	groupsCacheCtx    = context.Background()
	groupsCacheTTL    = 10 * time.Minute
)

func groupsCacheInit() {
	groupsCacheOnce.Do(func() {
		if environment.Settings.Cache.Backend != cache.BackendRedis {
			return
		}
		r := environment.Settings.Redis
		client := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", r.Host, r.Port),
			Username: r.Username,
			Password: r.Password,
			DB:       int(r.Database),
		})
		if err := client.Ping(groupsCacheCtx).Err(); err != nil {
			return
		}
		groupsCacheClient = client
	})
}

func groupsCacheKey(token string) string {
	return "quepasa:groups:" + strings.ToUpper(strings.TrimSpace(token))
}

// serveCachedGroups writes the cached groups payload and returns true on a cache hit.
// A "refresh=1"/"refresh=true" query param bypasses the cache.
func serveCachedGroups(w http.ResponseWriter, r *http.Request, token string) bool {
	refresh := r.URL.Query().Get("refresh")
	if refresh == "1" || strings.EqualFold(refresh, "true") {
		return false
	}
	groupsCacheInit()
	if groupsCacheClient == nil {
		return false
	}
	data, err := groupsCacheClient.Get(groupsCacheCtx, groupsCacheKey(token)).Bytes()
	if err != nil || len(data) == 0 {
		return false
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
	return true
}

// cacheGroupsBytes persists the serialized groups payload with a TTL.
func cacheGroupsBytes(token string, data []byte) {
	groupsCacheInit()
	if groupsCacheClient == nil || len(data) == 0 {
		return
	}
	_ = groupsCacheClient.Set(groupsCacheCtx, groupsCacheKey(token), data, groupsCacheTTL).Err()
}
