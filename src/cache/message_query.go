package cache

import (
	"sort"
	"strings"
)

const defaultQueryLimit = 50

// MessageQuery describes a paginated message lookup. Wid is required;
// ChatID and SinceTimestamp are optional filters.
type MessageQuery struct {
	Wid            string
	KeyPrefix      string // per-server isolation prefix (msgkey = "KEYPREFIX:ID"); empty = no prefix filter
	ChatID         string
	SinceTimestamp int64
	Page           int
	Limit          int
}

// FilterAndPaginate applies a MessageQuery to an already-listed set of entries.
// Backends without native paging (memory/redis/disk) reuse this. Newest first.
func FilterAndPaginate(entries []MessageRecordEntry, f MessageQuery) (items []MessageRecordEntry, total int) {
	var filtered []MessageRecordEntry
	for _, e := range entries {
		m := e.Record.Message
		if m == nil {
			continue
		}
		if f.KeyPrefix != "" && !strings.HasPrefix(e.Key, f.KeyPrefix+":") {
			continue
		}
		if f.Wid != "" && m.Wid != f.Wid {
			continue
		}
		if f.ChatID != "" && m.Chat.Id != f.ChatID {
			continue
		}
		if f.SinceTimestamp > 0 && m.Timestamp.Unix() < f.SinceTimestamp {
			continue
		}
		filtered = append(filtered, e)
	}
	sort.Slice(filtered, func(i, j int) bool {
		a, b := filtered[i].Record.Message, filtered[j].Record.Message
		if a.Timestamp.Equal(b.Timestamp) {
			return a.Id > b.Id // deterministic tie-break, newest-first
		}
		return a.Timestamp.After(b.Timestamp)
	})
	total = len(filtered)
	page, limit := f.Page, f.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultQueryLimit
	}
	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	return filtered[start:end], total
}
