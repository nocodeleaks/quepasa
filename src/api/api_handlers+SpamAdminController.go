package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/nocodeleaks/quepasa/library"
	"github.com/nocodeleaks/quepasa/models"
	"github.com/nocodeleaks/quepasa/runtime"
	"github.com/nocodeleaks/quepasa/whatsapp"
)

type spamSectionRequest struct {
	Token    string `json:"token"`
	Position int    `json:"position"`
	Enabled  *bool  `json:"enabled"`
	Label    string `json:"label"`
}

type spamSearchRequest struct {
	Search string `json:"search"`
	Limit  int    `json:"limit"`
}

type spamReorderRequest struct {
	Tokens []string `json:"tokens"`
}

type spamSectionView struct {
	Token     string `json:"token"`
	Wid       string `json:"wid,omitempty"`
	User      string `json:"user,omitempty"`
	ContextID string `json:"contextid,omitempty"`
	Verified  bool   `json:"verified"`
	Status    string `json:"status"`
	Ready     bool   `json:"ready"`

	InSpam   bool   `json:"inSpam"`
	Enabled  bool   `json:"enabled"`
	Position int    `json:"position,omitempty"`
	Label    string `json:"label,omitempty"`
}

func RegisterSpamAdminControllers(r chi.Router) {
	r.Get("/spam/status", SpamAdminStatusController)
	r.Get("/spam/sections", SpamAdminSectionsListController)
	r.Post("/spam/sections/search", SpamAdminSectionsSearchController)
	r.Post("/spam/sections", SpamAdminSectionUpsertController)
	r.Patch("/spam/sections", SpamAdminSectionUpsertController)
	r.Delete("/spam/sections", SpamAdminSectionDeleteController)
	r.Post("/spam/sections/reorder", SpamAdminSectionsReorderController)
}

func SpamAdminStatusController(w http.ResponseWriter, r *http.Request) {
	configured := isMasterKeyConfigured(models.ENV.MasterKey())
	RespondSuccess(w, map[string]interface{}{
		"configured": configured,
		"unlocked":   configured && IsMatchForMaster(r),
	})
}

func SpamAdminSectionsListController(w http.ResponseWriter, r *http.Request) {
	if !requireSpamMasterKey(w, r) {
		return
	}

	db := models.GetDatabase()
	sections, err := db.SpamSections.ListAll()
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	items := make([]spamSectionView, 0, len(sections))
	for _, section := range sections {
		if section == nil {
			continue
		}

		server, _ := db.Servers.FindByToken(section.Token)
		items = append(items, buildSpamSectionView(server, section))
	}

	RespondSuccess(w, map[string]interface{}{"items": items})
}

func SpamAdminSectionsSearchController(w http.ResponseWriter, r *http.Request) {
	if !requireSpamMasterKey(w, r) {
		return
	}

	request := spamSearchRequest{Limit: 100}
	if r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&request)
	}
	if value := strings.TrimSpace(library.GetRequestParameter(r, "search")); value != "" {
		request.Search = value
	}
	if value := strings.TrimSpace(library.GetRequestParameter(r, "limit")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			request.Limit = parsed
		}
	}
	if request.Limit <= 0 || request.Limit > 250 {
		request.Limit = 100
	}

	db := models.GetDatabase()
	spamMap, err := loadSpamSectionsMap(db.SpamSections)
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	needle := strings.ToLower(strings.TrimSpace(request.Search))
	servers := make([]*models.QpServer, 0)
	for _, server := range db.Servers.FindAll() {
		if server != nil {
			servers = append(servers, server)
		}
	}
	sort.SliceStable(servers, func(i, j int) bool {
		return strings.ToLower(servers[i].GetUser()) < strings.ToLower(servers[j].GetUser())
	})

	capacity := len(servers)
	if capacity > request.Limit {
		capacity = request.Limit
	}
	items := make([]spamSectionView, 0, capacity)
	for _, server := range servers {
		if needle != "" && !spamServerMatches(server, needle) {
			continue
		}

		items = append(items, buildSpamSectionView(server, spamMap[server.Token]))
		if len(items) >= request.Limit {
			break
		}
	}

	RespondSuccess(w, map[string]interface{}{"items": items})
}

func SpamAdminSectionUpsertController(w http.ResponseWriter, r *http.Request) {
	if !requireSpamMasterKey(w, r) {
		return
	}

	request := spamSectionRequest{}
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			RespondBadRequest(w, fmt.Errorf("invalid json body: %w", err))
			return
		}
	}
	if value := strings.TrimSpace(library.GetRequestParameter(r, "token")); value != "" {
		request.Token = value
	}

	request.Token = strings.TrimSpace(request.Token)
	if request.Token == "" {
		RespondBadRequest(w, fmt.Errorf("token is required"))
		return
	}

	db := models.GetDatabase()
	server, err := db.Servers.FindByToken(request.Token)
	if err != nil || server == nil {
		RespondNotFound(w, fmt.Errorf("server token not found"))
		return
	}

	enabled := true
	if request.Enabled != nil {
		enabled = *request.Enabled
	}

	section := &models.QpSpamSection{
		Token:    request.Token,
		Position: request.Position,
		Enabled:  enabled,
		Label:    request.Label,
	}
	if err := db.SpamSections.Upsert(section); err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	saved, _ := db.SpamSections.Find(section.Token)
	RespondSuccess(w, map[string]interface{}{
		"item": buildSpamSectionView(server, saved),
	})
}

func SpamAdminSectionDeleteController(w http.ResponseWriter, r *http.Request) {
	if !requireSpamMasterKey(w, r) {
		return
	}

	token := strings.TrimSpace(library.GetRequestParameter(r, "token"))
	if token == "" && r.ContentLength > 0 {
		request := spamSectionRequest{}
		_ = json.NewDecoder(r.Body).Decode(&request)
		token = strings.TrimSpace(request.Token)
	}
	if token == "" {
		RespondBadRequest(w, fmt.Errorf("token is required"))
		return
	}

	removed, err := models.GetDatabase().SpamSections.Delete(token)
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"removed": removed,
	})
}

func SpamAdminSectionsReorderController(w http.ResponseWriter, r *http.Request) {
	if !requireSpamMasterKey(w, r) {
		return
	}

	request := spamReorderRequest{}
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			RespondBadRequest(w, fmt.Errorf("invalid json body: %w", err))
			return
		}
	}

	store := models.GetDatabase().SpamSections
	position := 10
	for _, token := range request.Tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if err := store.UpdatePosition(token, position); err != nil {
			RespondErrorCode(w, err, http.StatusInternalServerError)
			return
		}
		position += 10
	}

	RespondSuccess(w, map[string]interface{}{"success": true})
}

func requireSpamMasterKey(w http.ResponseWriter, r *http.Request) bool {
	if !isMasterKeyConfigured(models.ENV.MasterKey()) {
		RespondErrorCode(w, fmt.Errorf("master key is not configured"), http.StatusForbidden)
		return false
	}
	if !IsMatchForMaster(r) {
		RespondErrorCode(w, errSpamMasterKeyRequired, http.StatusUnauthorized)
		return false
	}
	return true
}

func loadSpamSectionsMap(store models.QpDataSpamSectionsInterface) (map[string]*models.QpSpamSection, error) {
	result := map[string]*models.QpSpamSection{}
	sections, err := store.ListAll()
	if err != nil {
		return result, err
	}
	for _, section := range sections {
		if section != nil {
			result[section.Token] = section
		}
	}
	return result, nil
}

func buildSpamSectionView(server *models.QpServer, section *models.QpSpamSection) spamSectionView {
	view := spamSectionView{Status: whatsapp.Unknown.String()}
	if server != nil {
		view.Token = server.Token
		view.Wid = server.GetWId()
		view.User = server.GetUser()
		view.ContextID = server.GetContextId()
		view.Verified = server.Verified
	}
	if section != nil {
		view.Token = section.Token
		view.InSpam = true
		view.Enabled = section.Enabled
		view.Position = section.Position
		view.Label = section.Label
	}

	if live, ok := runtime.FindLiveSessionByToken(view.Token); ok && live != nil {
		status := live.GetStatus()
		view.Status = status.String()
		view.Ready = status == whatsapp.Ready
	}

	return view
}

func spamServerMatches(server *models.QpServer, needle string) bool {
	values := []string{
		server.Token,
		server.GetWId(),
		server.GetUser(),
		server.GetContextId(),
	}
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), needle) {
			return true
		}
	}
	return false
}
