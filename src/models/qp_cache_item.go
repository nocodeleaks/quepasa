package models

import "time"

type QpCacheItem struct {
	Key        string
	Value      interface{}
	Expiration time.Time
}

func (source *QpCacheItem) GetKey() string {
	return source.Key
}
