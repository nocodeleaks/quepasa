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

// SPASessionController returns the authenticated user session for SPA clients.
func SPASessionController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
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

// SPAServersController returns the user's servers, including disconnected records
// that still exist in the database.
func SPAServersController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

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

		items = append(items, BuildSPAServerSummary(dbServer, FindSPALiveServer(dbServer.Token)))
	}

	RespondSuccess(w, map[string]interface{}{
		"servers":  items,
		"total":    len(items),
		"version":  models.QpVersion,
		"username": user.Username,
	})
}

// SPAServersSearchController performs lightweight server-side filtering for SPA clients.
func SPAServersSearchController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

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

		liveServer := FindSPALiveServer(dbServer.Token)
		summary := BuildSPAServerSummary(dbServer, liveServer)

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

// SPAAccountController returns basic account information for the authenticated SPA user.
func SPAAccountController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	servers := runtime.ListLiveSessionsForUser(user.Username)
	RespondSuccess(w, map[string]interface{}{
		"user":         user,
		"serverCount":  len(servers),
		"version":      models.QpVersion,
		"hasMasterKey": len(strings.TrimSpace(models.ENV.MasterKey())) > 0,
	})
}

// SPAMasterKeyController keeps the legacy SPA route but never returns the master key secret.
func SPAMasterKeyController(w http.ResponseWriter, r *http.Request) {
	_, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	RespondSuccess(w, buildMasterKeyStatusResponse(strings.TrimSpace(models.ENV.MasterKey())))
}

// SPAServerInfoController returns server information for a token owned by the user.
func SPAServerInfoController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetSPATokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	dbServer, err := GetSPAOwnedServerRecord(user, token)
	if err != nil {
		if err.Error() == "server token not owned by user" {
			RespondErrorCode(w, err, http.StatusForbidden)
			return
		}
		RespondNotFound(w, err)
		return
	}

	liveServer := FindSPALiveServer(token)
	RespondSuccess(w, map[string]interface{}{
		"server": BuildSPAServerSummary(dbServer, liveServer),
	})
}

// SPAServerQRCodeController returns a QR code payload for a server that is not yet ready.
func SPAServerQRCodeController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetSPATokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	dbServer, err := GetSPAOwnedServerRecord(user, token)
	if err != nil {
		if err.Error() == "server token not owned by user" {
			RespondErrorCode(w, err, http.StatusForbidden)
			return
		}
		RespondNotFound(w, err)
		return
	}

	if liveServer := FindSPALiveServer(token); liveServer != nil && liveServer.GetStatus() == whatsapp.Ready {
		RespondSuccess(w, map[string]interface{}{
			"result":    "connected",
			"connected": true,
			"wid":       liveServer.Wid,
			"token":     token,
		})
		return
	}

	historySyncDays := parseSPAHistorySyncDays(r)

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

// SPAServerPairCodeController returns a phone pairing code for a server token owned by the user.
func SPAServerPairCodeController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetSPATokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	_, err = GetSPAOwnedServerRecord(user, token)
	if err != nil {
		if err.Error() == "server token not owned by user" {
			RespondErrorCode(w, err, http.StatusForbidden)
			return
		}
		RespondNotFound(w, err)
		return
	}

	if liveServer := FindSPALiveServer(token); liveServer != nil && liveServer.GetStatus() == whatsapp.Ready {
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

	historySyncDays := parseSPAHistorySyncDays(r)

	pairCode, err := runtime.PairSessionWithPhone(token, user.Username, phone, historySyncDays)
	if err != nil {
		RespondInterface(w, err)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"result":    "success",
		"pairCode":  pairCode,
		"formatted": formatSPAPairCode(pairCode),
	})
}

// SPAUsersListController returns the current authenticated user in a users collection shape.
func SPAUsersListController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
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

		isSelf := strings.EqualFold(current.Username, user.Username)
		items = append(items, map[string]interface{}{
			"username":   current.Username,
			"createdBy":  "",
			"created_by": "",
			"timestamp":  current.Timestamp.Format(time.RFC3339),
			"isSelf":     isSelf,
			"is_self":    isSelf,
		})
	}

	RespondSuccess(w, map[string]interface{}{
		"users": items,
	})
}

// SPAServerContactsController returns contacts for a live server owned by the user.
func SPAServerContactsController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetSPATokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	_, err = GetSPAOwnedServerRecord(user, token)
	if err != nil {
		if err.Error() == "server token not owned by user" {
			RespondErrorCode(w, err, http.StatusForbidden)
			return
		}
		RespondNotFound(w, err)
		return
	}

	server := FindSPALiveServer(token)
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

// SPAServerGroupsController returns joined groups for a live server owned by the user.
func SPAServerGroupsController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetSPATokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	_, err = GetSPAOwnedServerRecord(user, token)
	if err != nil {
		if err.Error() == "server token not owned by user" {
			RespondErrorCode(w, err, http.StatusForbidden)
			return
		}
		RespondNotFound(w, err)
		return
	}

	server := FindSPALiveServer(token)
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

// parseSPAHistorySyncDays extracts and validates the historysyncdays query parameter.
func parseSPAHistorySyncDays(r *http.Request) uint32 {
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

// formatSPAPairCode groups the raw pairing code for readability in the SPA.
func formatSPAPairCode(code string) string {
	if len(code) != 8 {
		return code
	}

	return code[0:2] + "-" + code[2:4] + "-" + code[4:6] + "-" + code[6:8]
}
