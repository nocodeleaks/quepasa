package environment

import (
	"strconv"
	"strings"
)

const (
	ENV_STORE_RETENTION_DAYS = "STORE_RETENTION_DAYS"
	ENV_DISPATCH_TYPES       = "DISPATCH_TYPES"
)

// MessageSettings holds instance-level (env) message config. RetentionDays:
// -1 = none (don't store), 0 = forever, N = keep N days. DispatchTypes: empty = all.
type MessageSettings struct {
	RetentionDays int             `json:"retention_days"`
	DispatchTypes map[string]bool `json:"dispatch_types"`
}

func parseStoreRetention(v string) int {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "0", "forever":
		return 0
	case "-1", "none":
		return -1
	}
	if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
		if n < 0 {
			return -1
		}
		return n
	}
	return 0
}

func parseDispatchTypes(v string) map[string]bool {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	set := map[string]bool{}
	for _, part := range strings.Split(v, ",") {
		t := strings.ToLower(strings.TrimSpace(part))
		if t != "" {
			set[t] = true
		}
	}
	if len(set) == 0 {
		return nil
	}
	return set
}

func NewMessageSettings() MessageSettings {
	return MessageSettings{
		RetentionDays: parseStoreRetention(getEnvOrDefaultString(ENV_STORE_RETENTION_DAYS, "")),
		DispatchTypes: parseDispatchTypes(getEnvOrDefaultString(ENV_DISPATCH_TYPES, "")),
	}
}
