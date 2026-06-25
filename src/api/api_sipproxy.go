package api

import (
	"net/http"

	log "github.com/nocodeleaks/quepasa/qplog"
	runtime "github.com/nocodeleaks/quepasa/runtime"
)

// AuthenticatedSIPProxyStatusController reports whether the SIP proxy is
// configured and whether it is currently running. Authenticated users only.
// The business logic lives in runtime.GetSIPProxyStatus; this controller only
// handles auth and the HTTP envelope.
func AuthenticatedSIPProxyStatusController(w http.ResponseWriter, r *http.Request) {
	_, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	status := runtime.GetSIPProxyStatus()
	log.Infof("sipproxy status query: configured=%t running=%t host=%s port=%d protocol=%s",
		status.Configured, status.Running, status.Host, status.Port, status.Protocol)

	RespondSuccess(w, status)
}
