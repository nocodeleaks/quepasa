package models

// Ordering by (Expiration) and then (Timestamp)
type QpCacheOrdering []QpCacheItem

func (source QpCacheOrdering) Len() int { return len(source) }

func (source QpCacheOrdering) Less(i, j int) bool {
	return source[i].Expiration.Before(source[j].Expiration)
}

func (source QpCacheOrdering) Swap(i, j int) {
	source[i], source[j] = source[j], source[i]
}
