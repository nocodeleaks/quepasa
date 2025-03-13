package library

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Cache struct {
	counter  atomic.Uint64
	cacheMap sync.Map
}

func (source *Cache) Count() uint64 {
	return source.counter.Load()
}

func (source *Cache) SetAny(key string, value interface{}, expiration time.Duration) {
	item := CacheItem{key, value, time.Now().Add(expiration)}
	source.SetCacheItem(item, "any")
}

// returns true if new item was appended
func (source *Cache) SetCacheItem(item CacheItem, from string) bool {
	_, loaded := source.cacheMap.Swap(item.Key, item)
	if loaded {
		return false
	} else {
		source.counter.Add(1)
	}

	return true
}

func (source *Cache) GetAny(key string) (interface{}, bool) {
	if val, ok := source.cacheMap.Load(key); ok {
		item := val.(CacheItem)
		if time.Now().Before(item.Expiration) {
			return item.Value, true
		} else {
			source.DeleteByKey(key)
		}
	}
	return nil, false
}

func (source *Cache) Delete(item CacheItem) {
	source.DeleteByKey(item.Key)
}

func (source *Cache) DeleteByKey(key string) {
	_, loaded := source.cacheMap.LoadAndDelete(key)
	if loaded {
		source.counter.Add(^uint64(0))
	}
}

// gets a copy as array of cached items
func (source *Cache) GetSliceOfCachedItems() (items []CacheItem) {
	source.cacheMap.Range(func(key, value any) bool {
		item := value.(CacheItem)
		items = append(items, item)
		return true
	})
	return items
}

// get a copy as array of cached items, ordered by expiration
func (source *Cache) GetOrdered() (items []CacheItem) {

	// filling array
	items = source.GetSliceOfCachedItems()

	// ordering
	sort.Sort(CacheOrdering(items))
	return
}

// remove old ones, by timestamp, until a maximum length
func (source *Cache) CleanUp(max uint64) {
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
