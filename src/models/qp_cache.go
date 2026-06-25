package models

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type QpCache struct {
	counter  atomic.Uint64
	cacheMap sync.Map
}

func (source *QpCache) Count() uint64 {
	return source.counter.Load()
}

func (source *QpCache) SetAny(key string, value interface{}, expiration time.Duration) {
	item := QpCacheItem{key, value, time.Now().Add(expiration)}
	source.SetCacheItem(item, "any")
}

// returns if it is a valid object, testing for now, it will not be necessary after debug
func (source *QpCache) SetCacheItem(item QpCacheItem, from string) bool {
	previous, loaded := source.cacheMap.Swap(item.Key, item)
	if loaded {
		return ValidateItemBecauseUNOAPIConflict(item, from, previous)
	} else {
		source.counter.Add(1)
	}

	return true
}

func (source *QpCache) GetAny(key string) (interface{}, bool) {
	if val, ok := source.cacheMap.Load(key); ok {
		item := val.(QpCacheItem)
		if time.Now().Before(item.Expiration) {
			return item.Value, true
		} else {
			source.DeleteByKey(key)
		}
	}
	return nil, false
}

func (source *QpCache) Delete(item QpCacheItem) {
	source.DeleteByKey(item.Key)
}

func (source *QpCache) DeleteByKey(key string) {
	_, loaded := source.cacheMap.LoadAndDelete(key)
	if loaded {
		source.counter.Add(^uint64(0))
	}
}

// gets a copy as array of cached items
func (source *QpCache) GetSliceOfCachedItems() (items []QpCacheItem) {
	source.cacheMap.Range(func(key, value any) bool {
		item := value.(QpCacheItem)
		items = append(items, item)
		return true
	})
	return items
}

// get a copy as array of cached items, ordered by expiration
func (source *QpCache) GetOrdered() (items []QpCacheItem) {

	// filling array
	items = source.GetSliceOfCachedItems()

	// ordering
	sort.Sort(QpCacheOrdering(items))
	return
}

// remove old ones, by timestamp, until a maximum length
func (source *QpCache) CleanUp(max uint64) {
	if max > 0 {

		// first checks only counter to avoid unecessary array searchs
		length := source.counter.Load()
		amount := length - max

		// if really has items to do a cleanup
		if amount > 0 {

			// searches the array for ordering and define the oldest items
			items := source.GetOrdered()

			// for thread safety, just in case that this cleaunp method is already running on another thread
			itemsLength := len(items)

			for i := 0; i < int(amount) && i < itemsLength; i++ {
				source.DeleteByKey(items[i].Key)
			}
		}
	}
}
