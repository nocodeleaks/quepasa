package library

// Ordering by (Expiration) and then (Timestamp)
type CacheOrdering []CacheItem

func (source CacheOrdering) Len() int { return len(source) }

func (source CacheOrdering) Less(i, j int) bool {
	return source[i].Expiration.Before(source[j].Expiration)
}

func (source CacheOrdering) Swap(i, j int) {
	source[i], source[j] = source[j], source[i]
}
