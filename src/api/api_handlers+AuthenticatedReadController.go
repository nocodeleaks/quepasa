package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	apiModels "github.com/nocodeleaks/quepasa/api/models"
	models "github.com/nocodeleaks/quepasa/models"
	runtime "github.com/nocodeleaks/quepasa/runtime"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	"github.com/skip2/go-qrcode"
)

// AuthenticatedSessionController returns the authenticated user session.
func AuthenticatedSessionController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"user": map[string]interface{}{
			"username": user.Username,
			"email":    user.Username,
		},
		"version": models.QpVersion,
	})
}

// AuthenticatedServersController returns the user's servers, including disconnected records
// that still exist in the database.
func AuthenticatedServersController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	scopedToken, hasScopedToken := getScopedSessionToken(r)

	dbServers, err := listPersistedServerRecords()
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}
	items := make([]map[string]interface{}, 0, len(dbServers))
	for _, dbServer := range dbServers {
		if dbServer == nil || dbServer.GetUser() != user.Username {
			continue
		}
		if hasScopedToken && !strings.EqualFold(dbServer.Token, scopedToken) {
			continue
		}

		items = append(items, BuildServerSummary(dbServer, FindLiveServer(dbServer.Token)))
	}

	RespondSuccess(w, map[string]interface{}{
		"servers":  items,
		"total":    len(items),
		"version":  models.QpVersion,
		"username": user.Username,
	})
}

// AuthenticatedServersSearchController performs lightweight server-side filtering.
func AuthenticatedServersSearchController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	scopedToken, hasScopedToken := getScopedSessionToken(r)

	var req struct {
		Query string `json:"query"`
		Token string `json:"token"`
		Phone string `json:"phone"`
		State string `json:"state"`
		Page  int    `json:"page"`
		Limit int    `json:"limit"`
	}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
			RespondErrorCode(w, err, http.StatusBadRequest)
			return
		}
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 25
	}

	query := strings.ToLower(strings.TrimSpace(req.Query))
	dbServers, err := listPersistedServerRecords()
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}
	items := make([]map[string]interface{}, 0, len(dbServers))
	for _, dbServer := range dbServers {
		if dbServer == nil || dbServer.GetUser() != user.Username {
			continue
		}
		if hasScopedToken && !strings.EqualFold(dbServer.Token, scopedToken) {
			continue
		}

		liveServer := FindLiveServer(dbServer.Token)
		summary := BuildServerSummary(dbServer, liveServer)

		tokenValue := strings.ToLower(dbServer.Token)
		widValue := strings.ToLower(dbServer.GetWId())
		stateValue := strings.ToLower(summary["state"].(string))

		match := true
		if req.Token != "" {
			match = match && strings.EqualFold(dbServer.Token, req.Token)
		}
		if req.Phone != "" {
			match = match && strings.Contains(widValue, strings.ToLower(req.Phone))
		}
		if req.State != "" {
			match = match && strings.Contains(stateValue, strings.ToLower(req.State))
		}
		if query != "" && !(strings.Contains(tokenValue, query) || strings.Contains(widValue, query) || strings.Contains(stateValue, query)) {
			match = false
		}

		if match {
			items = append(items, summary)
		}
	}

	total := len(items)
	start := (req.Page - 1) * req.Limit
	if start > total {
		start = total
	}
	end := start + req.Limit
	if end > total {
		end = total
	}

	RespondSuccess(w, map[string]interface{}{
		"items": items[start:end],
		"total": total,
		"page":  req.Page,
		"limit": req.Limit,
	})
}

// AuthenticatedAccountController returns basic account information for the authenticated user.
func AuthenticatedAccountController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	servers := runtime.ListLiveSessionsForUser(user.Username)
	if scopedToken, hasScopedToken := getScopedSessionToken(r); hasScopedToken {
		filtered := make([]*models.QpWhatsappServer, 0, 1)
		for _, server := range servers {
			if server == nil || !strings.EqualFold(server.Token, scopedToken) {
				continue
			}
			filtered = append(filtered, server)
		}
		servers = filtered
	}
	RespondSuccess(w, map[string]interface{}{
		"user":            user,
		"serverCount":     len(servers),
		"version":         models.QpVersion,
		"hasMasterKey":    len(strings.TrimSpace(models.ENV.MasterKey())) > 0,
		"relaxedSessions": isRelaxedSessions(),
	})
}

// AuthenticatedMasterKeyController keeps the legacy authenticated route but never returns the master key secret.
func AuthenticatedMasterKeyController(w http.ResponseWriter, r *http.Request) {
	_, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	RespondSuccess(w, buildMasterKeyStatusResponse(strings.TrimSpace(models.ENV.MasterKey())))
}

// AuthenticatedMasterVerifyController validates a master key candidate sent in the request body.
// Returns {"valid": true} when it matches the configured MASTERKEY, {"valid": false} otherwise.
// The configured key itself is never returned.
func AuthenticatedMasterVerifyController(w http.ResponseWriter, r *http.Request) {
	_, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	var body struct {
		Key string `json:"key"`
	}
	if r.Body == nil {
		RespondErrorCode(w, fmt.Errorf("missing request body"), http.StatusBadRequest)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondErrorCode(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	configured := strings.TrimSpace(models.ENV.MasterKey())
	if configured == "" {
		RespondSuccess(w, map[string]interface{}{"valid": false, "reason": "not_configured"})
		return
	}

	valid := strings.TrimSpace(body.Key) == configured
	RespondSuccess(w, map[string]interface{}{"valid": valid})
}

// AuthenticatedServerInfoController returns server information for a token owned by the user.
func AuthenticatedServerInfoController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	dbServer, err := GetOwnedServerRecord(user, token)
	if err != nil {
		if err.Error() == "server token not owned by user" {
			RespondErrorCode(w, err, http.StatusForbidden)
			return
		}
		RespondNotFound(w, err)
		return
	}

	liveServer := FindLiveServer(token)
	RespondSuccess(w, map[string]interface{}{
		"server": BuildServerSummary(dbServer, liveServer),
	})
}

// AuthenticatedServerQRCodeController returns a QR code payload for a server that is not yet ready.
func AuthenticatedServerQRCodeController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	dbServer, err := GetOwnedServerRecord(user, token)
	if err != nil {
		if err.Error() == "server token not owned by user" {
			RespondErrorCode(w, err, http.StatusForbidden)
			return
		}
		RespondNotFound(w, err)
		return
	}

	if liveServer := FindLiveServer(token); liveServer != nil && liveServer.GetStatus() == whatsapp.Ready {
		RespondSuccess(w, map[string]interface{}{
			"result":    "connected",
			"connected": true,
			"wid":       liveServer.Wid,
			"token":     token,
		})
		return
	}

	historySyncDays := parseHistorySyncDays(r)

	rawCode, err := runtime.GetSessionPairingQRCode(token, user.Username, historySyncDays)
	if err != nil {
		RespondInterface(w, err)
		return
	}

	png, err := qrcode.Encode(rawCode, qrcode.Medium, 256)
	if err != nil {
		RespondInterface(w, err)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"result":    "success",
		"connected": false,
		"token":     dbServer.Token,
		"qrcode":    "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
		"rawcode":   rawCode,
	})
}

// AuthenticatedServerPairCodeController returns a phone pairing code for a server token owned by the user.
func AuthenticatedServerPairCodeController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	_, err = GetOwnedServerRecord(user, token)
	if err != nil {
		if err.Error() == "server token not owned by user" {
			RespondErrorCode(w, err, http.StatusForbidden)
			return
		}
		RespondNotFound(w, err)
		return
	}

	if liveServer := FindLiveServer(token); liveServer != nil && liveServer.GetStatus() == whatsapp.Ready {
		RespondSuccess(w, map[string]interface{}{
			"result":    "connected",
			"connected": true,
			"wid":       liveServer.Wid,
			"token":     token,
		})
		return
	}

	phone := strings.TrimSpace(r.URL.Query().Get("phone"))
	if phone == "" {
		RespondErrorCode(w, errors.New("missing phone parameter"), http.StatusBadRequest)
		return
	}

	historySyncDays := parseHistorySyncDays(r)

	pairCode, err := runtime.PairSessionWithPhone(token, user.Username, phone, historySyncDays)
	if err != nil {
		RespondInterface(w, err)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"result":    "success",
		"pairCode":  pairCode,
		"formatted": formatPairCode(pairCode),
	})
}

// AuthenticatedUsersListController returns all users. Requires a valid X-Master-Key header.
func AuthenticatedUsersListController(w http.ResponseWriter, r *http.Request) {
	_, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	if !isMasterKeyRequest(r) {
		RespondErrorCode(w, fmt.Errorf("master key required"), http.StatusForbidden)
		return
	}

	users, err := runtime.ListPersistedUsers()
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	items := make([]map[string]interface{}, 0, len(users))
	for _, current := range users {
		if current == nil {
			continue
		}

		items = append(items, map[string]interface{}{
			"username":  current.Username,
			"timestamp": current.Timestamp.Format(time.RFC3339),
		})
	}

	RespondSuccess(w, map[string]interface{}{
		"users": items,
	})
}

// AuthenticatedServerContactsController returns contacts for a live server owned by the user.
func AuthenticatedServerContactsController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	_, err = GetOwnedServerRecord(user, token)
	if err != nil {
		if err.Error() == "server token not owned by user" {
			RespondErrorCode(w, err, http.StatusForbidden)
			return
		}
		RespondNotFound(w, err)
		return
	}

	server := FindLiveServer(token)
	if server == nil {
		RespondNotReady(w, fmt.Errorf("server is not active in memory"))
		return
	}

	contacts, err := server.GetContacts()
	if err != nil {
		RespondServerError(server, w, err)
		return
	}

	response := &apiModels.ContactsResponse{}
	response.Total = len(contacts)
	response.Contacts = contacts
	RespondSuccess(w, response)
}

// AuthenticatedServerGroupsController returns joined groups for a live server owned by the user.
func AuthenticatedServerGroupsController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	_, err = GetOwnedServerRecord(user, token)
	if err != nil {
		if err.Error() == "server token not owned by user" {
			RespondErrorCode(w, err, http.StatusForbidden)
			return
		}
		RespondNotFound(w, err)
		return
	}

	server := FindLiveServer(token)
	if server == nil {
		RespondNotReady(w, fmt.Errorf("server is not active in memory"))
		return
	}

	groups, err := server.GetGroupManager().GetJoinedGroups()
	if err != nil {
		RespondServerError(server, w, err)
		return
	}

	response := &apiModels.GroupsResponse{}
	response.Total = len(groups)
	response.Groups = groups
	RespondSuccess(w, response)
}

// parseHistorySyncDays extracts and validates the historysyncdays query parameter.
func parseHistorySyncDays(r *http.Request) uint32 {
	raw := strings.TrimSpace(r.URL.Query().Get("historysyncdays"))
	if raw == "" {
		return 0
	}
	value, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(value)
}

// formatPairCode groups the raw pairing code for readability in the SPA.
func formatPairCode(code string) string {
	if len(code) != 8 {
		return code
	}

	return code[0:2] + "-" + code[2:4] + "-" + code[4:6] + "-" + code[6:8]
}
