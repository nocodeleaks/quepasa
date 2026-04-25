package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
)

// restoreManualRequest is the payload expected by POST /restore/manual.
//
// Both fields are required. The token must already exist in the QuePasa
// database and the jid must be an active session in whatsmeow.sqlite.
type restoreManualRequest struct {
	// Token is the QuePasa server token (hex string stored in quepasa.sqlite).
	// Example: "mqUeLW3oMDFXR3m1finfu8rs"
	Token string `json:"token"`

	// JID is the full WhatsApp device identifier stored in whatsmeow.sqlite.
	// Example: "553176011595:18@s.whatsapp.net"
	JID string `json:"jid"`
}

// RestoreDiagnoseController is a GET /restore handler that returns a read-only
// diagnostic report showing which whatsmeow device sessions are orphaned (no
// matching QuePasa server with a wid) and which QuePasa servers are unlinked
// (wid is NULL in the database).
//
// Authentication: master key required.
//
// This endpoint never modifies any data — use POST /restore/auto or
// POST /restore/manual to apply actual changes.
//
// Response shape:
//
//	{
//	  "orphan_devices":    [ { "jid": "...", "phone": "...", "push_name": "..." } ],
//	  "unlinked_servers":  [ "token1", "token2" ],
//	  "restored":          [],
//	  "errors":            []
//	}
func RestoreDiagnoseController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Only administrators with the master key may trigger diagnostic operations
	// because they expose internal token and JID information.
	if !IsMatchForMaster(r) {
		RespondErrorCode(w, fmt.Errorf("master key required"), http.StatusUnauthorized)
		return
	}

	if models.WhatsappService == nil {
		RespondErrorCode(w, fmt.Errorf("whatsapp service not initialised"), http.StatusServiceUnavailable)
		return
	}

	report, err := models.WhatsappService.DiagnoseOrphaned()
	if err != nil {
		RespondErrorCode(w, fmt.Errorf("diagnosis failed: %v", err), http.StatusInternalServerError)
		return
	}

	RespondSuccess(w, report)
}

// RestoreAutoController is a POST /restore/auto handler that runs the automatic
// orphaned-account restore algorithm.
//
// Authentication: master key required.
//
// The algorithm works in two passes:
//  1. Phone-match: if a whatsmeow device JID's phone number matches a server
//     that is already in the runtime cache (e.g. loaded with an empty wid during
//     a previous session), that server is linked to the device.
//  2. One-to-one fallback: when exactly one orphan device and one unlinked server
//     remain after pass 1, the association is considered unambiguous and applied.
//
// Any ambiguous cases (multiple orphans, multiple servers, no phone match) are
// reported but left unchanged. Use POST /restore/manual for those.
//
// Response shape (same as /restore GET, with "restored" and "errors" populated):
//
//	{
//	  "orphan_devices":   [...],
//	  "unlinked_servers": [...],
//	  "restored":         [ { "token": "...", "jid": "..." } ],
//	  "errors":           [ { "token": "...", "jid": "...", "reason": "..." } ]
//	}
func RestoreAutoController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !IsMatchForMaster(r) {
		RespondErrorCode(w, fmt.Errorf("master key required"), http.StatusUnauthorized)
		return
	}

	if models.WhatsappService == nil {
		RespondErrorCode(w, fmt.Errorf("whatsapp service not initialised"), http.StatusServiceUnavailable)
		return
	}

	report, err := models.WhatsappService.RestoreOrphaned()
	if err != nil {
		RespondErrorCode(w, fmt.Errorf("auto restore failed: %v", err), http.StatusInternalServerError)
		return
	}

	RespondSuccess(w, report)
}

// RestoreManualController is a POST /restore/manual handler that links a
// specific QuePasa server token to a specific whatsmeow device JID.
//
// Authentication: master key required.
//
// Use this endpoint when the automatic restore cannot disambiguate between
// multiple orphan devices or unlinked servers.
//
// Request body (JSON):
//
//	{
//	  "token": "mqUeLW3oMDFXR3m1finfu8rs",
//	  "jid":   "553176011595:18@s.whatsapp.net"
//	}
//
// On success the server record is updated in the database and the runtime cache
// is reloaded, making the connection immediately available without a restart.
//
// Response on success (HTTP 200):
//
//	{ "result": "ok", "token": "...", "jid": "..." }
//
// Response on error (HTTP 400 / 500):
//
//	{ "result": "<error message>" }
func RestoreManualController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !IsMatchForMaster(r) {
		RespondErrorCode(w, fmt.Errorf("master key required"), http.StatusUnauthorized)
		return
	}

	if models.WhatsappService == nil {
		RespondErrorCode(w, fmt.Errorf("whatsapp service not initialised"), http.StatusServiceUnavailable)
		return
	}

	// Read and parse the request body.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		RespondErrorCode(w, fmt.Errorf("failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	var req restoreManualRequest
	if err := json.Unmarshal(body, &req); err != nil {
		RespondErrorCode(w, fmt.Errorf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields before handing off to the service.
	if len(req.Token) == 0 {
		RespondErrorCode(w, fmt.Errorf("field 'token' is required"), http.StatusBadRequest)
		return
	}
	if len(req.JID) == 0 {
		RespondErrorCode(w, fmt.Errorf("field 'jid' is required"), http.StatusBadRequest)
		return
	}

	if err := models.WhatsappService.RestoreManual(req.Token, req.JID); err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	RespondSuccess(w, map[string]string{
		"result": "ok",
		"token":  req.Token,
		"jid":    req.JID,
	})
}
