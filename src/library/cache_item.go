package library

import "time"

type CacheItem struct {
	Key        string
	Value      interface{}
	Expiration time.Time
}

func (source *CacheItem) GetKey() string {
	return source.Key
}
