package whatsmeow

import (
	"time"

	library "github.com/nocodeleaks/quepasa/library"
)

const DEFAULTEXPIRATION_WGIC time.Duration = time.Duration(1 * time.Hour)

type WhatsmeowGroupInfoCache struct {
	library.Cache
}

func GetCacheExpiration() time.Time {
	return time.Now().Add(DEFAULTEXPIRATION_WGIC)
}

func (source *WhatsmeowGroupInfoCache) Append(id string, title string, from string) bool {
	expiration := GetCacheExpiration()
	item := library.CacheItem{
		Key:        id,
		Value:      title,
		Expiration: expiration,
	}
	return source.SetCacheItem(item, "groupinfo-"+from)
}

func (source *WhatsmeowGroupInfoCache) Get(id string) (title string) {
	cached, success := source.GetAny(id)
	if success {
		title, _ = cached.(string)
	}
	return
}

var GroupInfoCache WhatsmeowGroupInfoCache = WhatsmeowGroupInfoCache{}
