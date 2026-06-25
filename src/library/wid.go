/*
	Methods related to WId (WhatsApp ID)
*/

package library

import (
	"strings"
)

func TrimSessionIdFromWIdString(wid string) string {
	if strings.Contains(wid, ":") {
		prefix := strings.Split(wid, ":")[0]
		suffix := strings.Split(wid, "@")[1]
		return prefix + "@" + suffix
	}
	return wid
}
