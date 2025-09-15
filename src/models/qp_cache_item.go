package models

import "time"

type QpCacheItem struct {
	Key        string
	Value      any
	Expiration time.Time
}

func (source *QpCacheItem) GetKey() string {
	return source.Key
}
