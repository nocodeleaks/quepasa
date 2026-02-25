package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	websocket "github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	qrcode "github.com/skip2/go-qrcode"

	environment "github.com/nocodeleaks/quepasa/environment"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"
)

// parseWhatsappBoolean converts JSON value to WhatsappBoolean
// Accepts: -1, 0, 1 (numbers) or "true", "false", null
func parseWhatsappBoolean(v interface{}) (whatsapp.WhatsappBoolean, error) {
	switch val := v.(type) {
	case float64:
		return whatsapp.WhatsappBoolean(int(val)), nil
	case int:
		return whatsapp.WhatsappBoolean(val), nil
	case bool:
		if val {
			return whatsapp.TrueBooleanType, nil
		}
		return whatsapp.FalseBooleanType, nil
	case nil:
		return whatsapp.UnSetBooleanType, nil
	case string:
		switch strings.ToLower(strings.TrimSpace(val)) {
		case "1", "true", "yes":
			return whatsapp.TrueBooleanType, nil
		case "-1", "false", "no":
			return whatsapp.FalseBooleanType, nil
		case "0", "":
			return whatsapp.UnSetBooleanType, nil
		}
	}
	return whatsapp.UnSetBooleanType, fmt.Errorf("invalid WhatsappBoolean value: %v", v)
}

// SPASessionController returns basic session info for SPA
// GET /api/session
func SPASessionController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		if err == models.ErrFormUnauthenticated {
			RespondUnauthorized(w, err)
			return
		}
		RespondInterface(w, err)
		return
	}

	resp := map[string]interface{}{
		"user": map[string]interface{}{
			"username": user.Username,
			"email":    user.Username,
			"level":    "",
		},
		"version":         models.QpVersion,
		"serversViewMode": environment.Settings.Form.ServersViewMode,
		"branding":        environment.Settings.Branding,
	}

	RespondInterface(w, resp)
}

// SPAServersController returns the list of servers for the authenticated user
// GET /api/servers
func SPAServersController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		if err == models.ErrFormUnauthenticated {
			RespondUnauthorized(w, err)
			return
		}
		RespondInterface(w, err)
		return
	}

	// Fetch servers from database (includes disconnected ones)
	isAdmin := false
	allServers := models.WhatsappService.DB.Servers.FindAll()
	dbServers := make([]*models.QpServer, 0)
	for _, server := range allServers {
		if server != nil && server.User == user.Username {
			dbServers = append(dbServers, server)
		}
	}

	var dbErr error
	if dbErr != nil {
		log.Errorf("Error fetching servers from database: %v", dbErr)
		RespondServerError(nil, w, dbErr)
		return
	}

	items := make([]map[string]interface{}, 0, len(dbServers))

	// Helper function to convert whatsapp.WhatsappBoolean to bool
	toBool := func(wb whatsapp.WhatsappBoolean) bool {
		return wb == whatsapp.TrueBooleanType
	}

	for _, dbServer := range dbServers {
		// Try to get live server from memory for real-time state
		liveServer, _ := models.WhatsappService.FindByToken(dbServer.Token)

		var state whatsapp.WhatsappConnectionState
		var timestamps models.QpTimestamps
		var dispatchCount, webhookCount, rabbitmqCount int

		if liveServer != nil {
			// Server is in memory - use live data
			state = liveServer.GetState()
			timestamps = liveServer.Timestamps

			dispatchings := liveServer.GetDispatchingByFilter("")
			dispatchCount = len(dispatchings)

			webhooks := liveServer.GetWebhooks()
			webhookCount = len(webhooks)

			rabbitConfigs := liveServer.GetRabbitMQConfigsByQueue("")
			rabbitmqCount = len(rabbitConfigs)
		} else {
			// Server is not in memory - use database data with offline state
			state = whatsapp.Disconnected
		}

		uptimeSeconds := int64(0)
		if !timestamps.Start.IsZero() {
			uptimeSeconds = int64(time.Since(timestamps.Start).Seconds())
		}

		// Get connection status string
		connectionStatus := "Unverified"
		if state == whatsapp.Ready {
			connectionStatus = "Ready"
		} else if state == whatsapp.Connecting || state == whatsapp.Starting {
			connectionStatus = "Connecting"
		} else if state == whatsapp.Disconnected {
			connectionStatus = "Disconnected"
		} else if state == whatsapp.UnVerified {
			connectionStatus = "Unverified"
		}

		// Include owner info for admin users
		serverItem := map[string]interface{}{
			"token":            dbServer.Token,
			"wid":              dbServer.Wid,
			"state":            state.String(),
			"state_code":       state.EnumIndex(),
			"reconnect":        false, // Not available from DB
			"start_time":       timestamps.Start,
			"last_update":      timestamps.Update,
			"uptime_seconds":   uptimeSeconds,
			"verified":         dbServer.Verified,
			"devel":            dbServer.Devel,
			"has_webhooks":     webhookCount > 0,
			"has_websockets":   false,
			"webhook_count":    webhookCount,
			"rabbitmq_count":   rabbitmqCount,
			"dispatch_count":   dispatchCount,
			"has_dispatching":  dispatchCount > 0,
			"connection":       connectionStatus,
			"groups":           toBool(dbServer.Groups),
			"broadcasts":       toBool(dbServer.Broadcasts),
			"read_receipts":    toBool(dbServer.ReadReceipts),
			"calls":            toBool(dbServer.Calls),
		}

		// Add owner info for admin users viewing other users' servers
		if isAdmin && dbServer.User != user.Username {
			serverItem["owner"] = dbServer.User
		}

		items = append(items, serverItem)
	}

	resp := map[string]interface{}{
		"result":          "success",
		"version":         models.QpVersion,
		"servers":         items,
		"serversViewMode": environment.Settings.Form.ServersViewMode,
		"isAdmin":         isAdmin,
	}

	RespondInterface(w, resp)
}

// SPAServersSearchController searches servers for the authenticated user
// POST /api/servers/search
func SPAServersSearchController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		if err == models.ErrFormUnauthenticated {
			RespondUnauthorized(w, err)
			return
		}
		RespondInterface(w, err)
		return
	}

	// Parse request body
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

	servers := models.GetServersForUser(user)
	items := make([]map[string]interface{}, 0, len(servers))

	// Simple filter function
	q := strings.TrimSpace(strings.ToLower(req.Query))
	for _, server := range servers {
		// Ownership guard (already by GetServersForUser)
		// Build searchable fields
		token := strings.ToLower(server.Token)
		wid := strings.ToLower(server.Wid)
		state := strings.ToLower(server.GetState().String())

		// Match logic
		match := true
		if len(req.Token) > 0 {
			// exact match by token
			match = match && strings.EqualFold(server.Token, req.Token)
		}
		if len(req.Phone) > 0 {
			phone := strings.ToLower(req.Phone)
			match = match && strings.Contains(wid, phone)
		}
		if len(req.State) > 0 {
			match = match && strings.Contains(state, strings.ToLower(req.State))
		}
		if len(q) > 0 {
			if !(strings.Contains(token, q) || strings.Contains(wid, q) || strings.Contains(state, q)) {
				match = false
			}
		}

		if !match {
			continue
		}

		stateVal := server.GetState()
		timestamps := server.Timestamps
		uptimeSeconds := int64(0)
		if !timestamps.Start.IsZero() {
			uptimeSeconds = int64(time.Since(timestamps.Start).Seconds())
		}

		dispatchings := server.GetDispatchingByFilter("")
		dispatchCount := len(dispatchings)
		webhooks := server.GetWebhooks()
		rabbitConfigs := server.GetRabbitMQConfigsByQueue("")
		rabbitmqCount := len(rabbitConfigs)

		connectionStatus := "Unverified"
		if stateVal == whatsapp.Ready {
			connectionStatus = "Ready"
		} else if stateVal == whatsapp.Connecting || stateVal == whatsapp.Starting {
			connectionStatus = "Connecting"
		} else if stateVal == whatsapp.Disconnected {
			connectionStatus = "Disconnected"
		} else if stateVal == whatsapp.UnVerified {
			connectionStatus = "Unverified"
		}

		items = append(items, map[string]interface{}{
			"token":           server.Token,
			"wid":             server.Wid,
			"state":           stateVal.String(),
			"uptime_seconds":  uptimeSeconds,
			"verified":        server.Verified,
			"devel":           server.Devel,
			"webhook_count":   len(webhooks),
			"rabbitmq_count":  rabbitmqCount,
			"dispatch_count":  dispatchCount,
			"has_dispatching": dispatchCount > 0,
			"connection":      connectionStatus,
		})
	}

	// Pagination
	total := len(items)
	start := (req.Page - 1) * req.Limit
	end := start + req.Limit
	if start < 0 {
		start = 0
	}
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	paged := items[start:end]

	resp := map[string]interface{}{
		"result": "success",
		"total":  total,
		"page":   req.Page,
		"limit":  req.Limit,
		"servers": paged,
	}

	RespondInterface(w, resp)
}

// SPAAccountController returns account info for SPA
// GET /api/account
func SPAAccountController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		if err == models.ErrFormUnauthenticated {
			RespondErrorCode(w, err, http.StatusUnauthorized)
			return
		}
		RespondInterface(w, err)
		return
	}

	data := map[string]interface{}{
		"user":         user,
		"servers":      models.GetServersForUser(user),
		"version":      models.QpVersion,
		"hasMasterKey": len(environment.Settings.API.MasterKey) > 0,
	}

	RespondSuccess(w, data)
}

// SPAMasterKeyController returns the master key for authenticated users
// GET /api/account/masterkey
func SPAMasterKeyController(w http.ResponseWriter, r *http.Request) {
	_, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	master := environment.Settings.API.MasterKey
	if len(master) == 0 {
		RespondErrorCode(w, errors.New("master key not configured"), http.StatusNotFound)
		return
	}

	RespondSuccess(w, map[string]interface{}{"masterKey": master})
}

// SPAWebHooksController returns webhooks for a server
// GET /api/webhooks?token=TOKEN
func SPAWebHooksController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token := r.URL.Query().Get("token")
	if len(token) == 0 {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondErrorCode(w, gerr, http.StatusNotFound)
		return
	}

	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	response := map[string]interface{}{
		"server":   server,
		"webhooks": server.GetWebhooks(),
	}
	RespondSuccess(w, response)
}

// SPAWebHooksCreateController creates a webhook
// POST /api/webhooks
func SPAWebHooksCreateController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	var webhook struct {
		Token           string      `json:"token"`
		Url             string      `json:"url"`
		TrackId         string      `json:"trackId"`
		ForwardInternal bool        `json:"forwardInternal"`
		Broadcasts      bool        `json:"broadcasts"`
		Groups          bool        `json:"groups"`
		ReadReceipts    bool        `json:"readReceipts"`
		Calls           bool        `json:"calls"`
		Extra           interface{} `json:"extra"`
	}
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	token := webhook.Token
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	if token == "" {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondErrorCode(w, gerr, http.StatusNotFound)
		return
	}
	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	boolToWB := func(v bool) whatsapp.WhatsappBoolean {
		if v {
			return whatsapp.TrueBooleanType
		}
		return whatsapp.FalseBooleanType
	}

	d := &models.QpDispatching{
		ConnectionString: webhook.Url,
		Type:             models.DispatchingTypeWebhook,
		ForwardInternal:  webhook.ForwardInternal,
		TrackId:          webhook.TrackId,
		Extra:            webhook.Extra,
	}
	d.WhatsappOptions = whatsapp.WhatsappOptions{
		ReadReceipts: boolToWB(webhook.ReadReceipts),
		Groups:       boolToWB(webhook.Groups),
		Broadcasts:   boolToWB(webhook.Broadcasts),
		Calls:        boolToWB(webhook.Calls),
	}
	affected, derr := server.DispatchingAddOrUpdate(d)
	if derr != nil {
		RespondInterface(w, derr)
		return
	}

	RespondSuccess(w, map[string]interface{}{"affected": affected})
}

// SPAWebHooksDeleteController deletes a webhook
// DELETE /api/webhooks
func SPAWebHooksDeleteController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.FormValue("token")
	}
	if token == "" {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	url := r.FormValue("url")
	if url == "" {
		var body struct{ Url string `json:"url"` }
		_ = json.NewDecoder(r.Body).Decode(&body)
		url = body.Url
	}
	if url == "" {
		RespondErrorCode(w, errors.New("missing url parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondErrorCode(w, gerr, http.StatusNotFound)
		return
	}
	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	affected, derr := server.DispatchingRemove(url)
	if derr != nil {
		RespondInterface(w, derr)
		return
	}

	RespondSuccess(w, map[string]interface{}{"affected": affected})
}

// SPAWebHooksUpdateController updates webhook fields (url, context, trackid, extra, options)
// PUT /api/webhooks
func SPAWebHooksUpdateController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	var webhook struct {
		Token           string      `json:"token"`
		OriginalUrl     string      `json:"originalUrl"`     // URL original para identificar o webhook
		Url             string      `json:"url"`             // Nova URL (pode ser igual à original)
		TrackId         *string     `json:"trackId"`         // Ponteiro para permitir null/undefined
		ForwardInternal *bool       `json:"forwardInternal"` // Ponteiro para permitir null/undefined
		Broadcasts      *int        `json:"broadcasts"`      // -1, 0, 1
		Groups          *int        `json:"groups"`          // -1, 0, 1
		ReadReceipts    *int        `json:"readReceipts"`    // -1, 0, 1
		Calls           *int        `json:"calls"`           // -1, 0, 1
		Extra           interface{} `json:"extra"`
	}
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	token := webhook.Token
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	if token == "" {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	if webhook.OriginalUrl == "" {
		RespondErrorCode(w, errors.New("missing originalUrl parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondErrorCode(w, gerr, http.StatusNotFound)
		return
	}
	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	// Se a URL mudou, precisamos remover a antiga e criar uma nova
	if webhook.Url != "" && webhook.Url != webhook.OriginalUrl {
		// Remove a antiga
		_, derr := server.DispatchingRemove(webhook.OriginalUrl)
		if derr != nil {
			RespondInterface(w, derr)
			return
		}
	}

	// Prepara o dispatching atualizado
	newUrl := webhook.Url
	if newUrl == "" {
		newUrl = webhook.OriginalUrl
	}

	d := &models.QpDispatching{
		ConnectionString: newUrl,
		Type:             models.DispatchingTypeWebhook,
		Extra:            webhook.Extra,
	}

	// Set optional fields
	if webhook.TrackId != nil {
		d.TrackId = *webhook.TrackId
	}
	if webhook.ForwardInternal != nil {
		d.ForwardInternal = *webhook.ForwardInternal
	}

	// Handle tri-state boolean options
	if webhook.Broadcasts != nil {
		d.Broadcasts = whatsapp.WhatsappBoolean(*webhook.Broadcasts)
	}
	if webhook.Groups != nil {
		d.Groups = whatsapp.WhatsappBoolean(*webhook.Groups)
	}
	if webhook.ReadReceipts != nil {
		d.ReadReceipts = whatsapp.WhatsappBoolean(*webhook.ReadReceipts)
	}
	if webhook.Calls != nil {
		d.Calls = whatsapp.WhatsappBoolean(*webhook.Calls)
	}

	affected, derr := server.DispatchingAddOrUpdate(d)
	if derr != nil {
		RespondInterface(w, derr)
		return
	}

	RespondSuccess(w, map[string]interface{}{"affected": affected})
}

// SPARabbitMQController returns RabbitMQ configs for a server
// GET /api/rabbitmq?token=TOKEN
func SPARabbitMQController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token := r.URL.Query().Get("token")
	if len(token) == 0 {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondErrorCode(w, gerr, http.StatusNotFound)
		return
	}

	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	response := map[string]interface{}{
		"server":   server,
		"rabbitmq": server.GetRabbitMQConfigsByQueue(""),
	}
	RespondSuccess(w, response)
}

// SPARabbitMQCreateController creates a RabbitMQ config
// POST /api/rabbitmq
func SPARabbitMQCreateController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	var cfg struct {
		Token            string      `json:"token"`
		ConnectionString string      `json:"connectionString"`
		TrackId          string      `json:"trackId"`
		ReadReceipts     bool        `json:"readReceipts"`
		Groups           bool        `json:"groups"`
		Broadcasts       bool        `json:"broadcasts"`
		Extra            interface{} `json:"extra"`
	}
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	token := cfg.Token
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	if token == "" {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondErrorCode(w, gerr, http.StatusNotFound)
		return
	}
	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	boolToWB := func(v bool) whatsapp.WhatsappBoolean {
		if v {
			return whatsapp.TrueBooleanType
		}
		return whatsapp.FalseBooleanType
	}

	d := &models.QpDispatching{
		ConnectionString: cfg.ConnectionString,
		Type:             models.DispatchingTypeRabbitMQ,
		TrackId:          cfg.TrackId,
		Extra:            cfg.Extra,
	}
	d.WhatsappOptions = whatsapp.WhatsappOptions{
		ReadReceipts: boolToWB(cfg.ReadReceipts),
		Groups:       boolToWB(cfg.Groups),
		Broadcasts:   boolToWB(cfg.Broadcasts),
	}
	affected, derr := server.DispatchingAddOrUpdate(d)
	if derr != nil {
		RespondInterface(w, derr)
		return
	}

	RespondSuccess(w, map[string]interface{}{"affected": affected})
}

// SPARabbitMQDeleteController deletes a RabbitMQ config
// DELETE /api/rabbitmq
func SPARabbitMQDeleteController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.FormValue("token")
	}
	if token == "" {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	cs := r.FormValue("connectionString")
	if cs == "" {
		var body struct{ ConnectionString string `json:"connectionString"` }
		_ = json.NewDecoder(r.Body).Decode(&body)
		cs = body.ConnectionString
	}
	if cs == "" {
		RespondErrorCode(w, errors.New("missing connectionString parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondErrorCode(w, gerr, http.StatusNotFound)
		return
	}
	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	affected, derr := server.DispatchingRemove(cs)
	if derr != nil {
		RespondInterface(w, derr)
		return
	}

	RespondSuccess(w, map[string]interface{}{"affected": affected})
}

// SPAServerInfoController returns server info
// GET /api/server/{token}/info
func SPAServerInfoController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token := chi.URLParam(r, "token")
	if len(token) == 0 {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondInterface(w, gerr)
		return
	}

	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	// Get current state
	state := server.GetState()

	RespondSuccess(w, map[string]interface{}{
		"server":    server,
		"state":     state.String(),
		"stateCode": state.EnumIndex(),
		"connected": state == whatsapp.Ready,
		"wid":       server.Wid,
	})
}

// SPAServerQRCodeController returns QR code for server connection
// GET /api/server/{token}/qrcode
func SPAServerQRCodeController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token := chi.URLParam(r, "token")
	if len(token) == 0 {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondInterface(w, gerr)
		return
	}

	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	// Check if already connected
	if server.GetStatus() == whatsapp.Ready {
		RespondSuccess(w, map[string]interface{}{
			"result":    "connected",
			"connected": true,
			"wid":       server.Wid,
		})
		return
	}

	// Get pairing info
	pairing := &models.QpWhatsappPairing{
		Username: user.Username,
		Token:    token,
	}

	// Apply history sync from environment default when available
	if d := models.ENV.HistorySync(); d != nil {
		pairing.HistorySyncDays = *d
	}

	con, cerr := pairing.GetConnection()
	if cerr != nil {
		RespondInterface(w, cerr)
		return
	}

	// Get QR code string
	qrCodeStr := con.GetWhatsAppQRCode()
	if qrCodeStr == "" {
		RespondErrorCode(w, errors.New("empty QR code - server may already be connected"), http.StatusBadRequest)
		return
	}

	// Generate QR code image using skip2/go-qrcode
	png, qerr := qrcode.Encode(qrCodeStr, qrcode.Medium, 256)
	if qerr != nil {
		RespondInterface(w, qerr)
		return
	}

	// Return as base64 image
	RespondSuccess(w, map[string]interface{}{
		"result":  "success",
		"qrcode":  "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
		"rawcode": qrCodeStr,
	})
}

// SPAServerPairCodeController returns pair code for server connection
// GET /api/server/{token}/paircode
func SPAServerPairCodeController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token := chi.URLParam(r, "token")
	if len(token) == 0 {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondInterface(w, gerr)
		return
	}

	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	// Check if already connected
	if server.GetStatus() == whatsapp.Ready {
		RespondSuccess(w, map[string]interface{}{
			"result":    "connected",
			"connected": true,
			"wid":       server.Wid,
		})
		return
	}

	// Get phone from query
	phone := r.URL.Query().Get("phone")
	if phone == "" {
		RespondErrorCode(w, errors.New("missing phone parameter"), http.StatusBadRequest)
		return
	}

	// Get pairing info
	pairing := &models.QpWhatsappPairing{
		Username: user.Username,
		Token:    token,
	}

	// Apply history sync from environment default when available
	if d := models.ENV.HistorySync(); d != nil {
		pairing.HistorySyncDays = *d
	}

	con, cerr := pairing.GetConnection()
	if cerr != nil {
		RespondInterface(w, cerr)
		return
	}

	// Get pair code using PairPhone method
	pairCode, pcerr := con.PairPhone(phone)
	if pcerr != nil {
		if strings.Contains(pcerr.Error(), "missing <link_code_pairing_wrapped_primary_ephemeral_pub>") {
			// Return a friendly error to SPA with suggestion to retry
			RespondErrorCode(w, fmt.Errorf("pairing failed due to incomplete notification from WhatsApp: %s", pcerr.Error()), http.StatusServiceUnavailable)
			return
		}
		RespondInterface(w, pcerr)
		return
	}

	// Format pair code as XX-XX-XX-XX for readability
	formattedCode := formatPairCode(pairCode)

	RespondSuccess(w, map[string]interface{}{
		"result":    "success",
		"paircode":  pairCode,
		"formatted": formattedCode,
	})
}

// formatPairCode formats pair code as XX-XX-XX-XX
func formatPairCode(code string) string {
	if len(code) != 8 {
		return code
	}
	return code[0:2] + "-" + code[2:4] + "-" + code[4:6] + "-" + code[6:8]
}

// SPAServerCreateController creates a new server
// POST /api/server/create
func SPAServerCreateController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	// Generate new token
	token := uuid.New().String()

	// Create server info with default options from environment
	info := &models.QpServer{
		Token: token,
		User:  user.Username,
	}

	// Apply default WhatsApp options from environment
	// Convert WhatsappBooleanExtended to WhatsappBoolean for storage
	info.Groups = models.ENV.Groups().ToWhatsappBoolean()
	info.Broadcasts = models.ENV.Broadcasts().ToWhatsappBoolean()
	info.ReadReceipts = models.ENV.ReadReceipts().ToWhatsappBoolean()
	info.Calls = models.ENV.Calls().ToWhatsappBoolean()
	// ReadUpdate returns bool, convert to WhatsappBoolean
	if models.ENV.ReadUpdate() {
		info.ReadUpdate = whatsapp.TrueBooleanType
	} else {
		info.ReadUpdate = whatsapp.FalseBooleanType
	}

	// Append new server
	server, cerr := models.WhatsappService.AppendNewServer(info)
	if cerr != nil {
		RespondInterface(w, cerr)
		return
	}

	// Persist server to database so it survives restarts
	if saveErr := server.Save("server created via SPA"); saveErr != nil {
		// Remove from in-memory cache if save fails
		delete(models.WhatsappService.Servers, server.Token)
		RespondInterface(w, saveErr)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"server": server,
		"token":  server.Token,
	})
}

// SPAServerDeleteController deletes a server
// POST /api/delete
func SPAServerDeleteController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	// Get token from body or form
	r.ParseForm()
	token := r.Form.Get("token")
	if token == "" {
		var body struct{ Token string `json:"token"` }
		_ = json.NewDecoder(r.Body).Decode(&body)
		token = body.Token
	}
	if token == "" {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondErrorCode(w, gerr, http.StatusNotFound)
		return
	}

	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	// Delete the server
	derr := models.WhatsappService.Delete(server, "server deleted via SPA")
	if derr != nil {
		RespondInterface(w, derr)
		return
	}

	RespondSuccess(w, map[string]interface{}{"result": "success"})
}

// SPAServerDebugController toggles debug mode for a server
// POST /api/debug
func SPAServerDebugController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	// Get token from body or form
	r.ParseForm()
	token := r.Form.Get("token")
	if token == "" {
		var body struct{ Token string `json:"token"` }
		_ = json.NewDecoder(r.Body).Decode(&body)
		token = body.Token
	}
	if token == "" {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondErrorCode(w, gerr, http.StatusNotFound)
		return
	}

	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	// Toggle debug mode
	_, terr := server.ToggleDevel()
	if terr != nil {
		RespondInterface(w, terr)
		return
	}

	RespondSuccess(w, map[string]interface{}{"devel": server.Devel})
}

// SPAServerUpdateController updates server options (history sync etc)
// POST /api/server/{token}/update
func SPAServerUpdateController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token := chi.URLParam(r, "token")
	if len(token) == 0 {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondInterface(w, gerr)
		return
	}

	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	// Parse body (accept form or JSON)
	r.ParseForm()
	var jb map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&jb)

	updated := false

	// Handle WhatsApp options (groups, broadcasts, readreceipts, calls, readupdate)
	// Values: -1 = false, 0 = unset, 1 = true
	whatsappOptions := server.GetOptions()

	if v, ok := jb["groups"]; ok {
		if val, err := parseWhatsappBoolean(v); err == nil {
			whatsappOptions.Groups = val
			updated = true
		}
	}
	if v, ok := jb["broadcasts"]; ok {
		if val, err := parseWhatsappBoolean(v); err == nil {
			whatsappOptions.Broadcasts = val
			updated = true
		}
	}
	if v, ok := jb["readreceipts"]; ok {
		if val, err := parseWhatsappBoolean(v); err == nil {
			whatsappOptions.ReadReceipts = val
			updated = true
		}
	}
	if v, ok := jb["calls"]; ok {
		if val, err := parseWhatsappBoolean(v); err == nil {
			whatsappOptions.Calls = val
			updated = true
		}
	}
	if v, ok := jb["readupdate"]; ok {
		if val, err := parseWhatsappBoolean(v); err == nil {
			whatsappOptions.ReadUpdate = val
			updated = true
		}
	}

	if !updated {
		RespondSuccess(w, map[string]interface{}{"server": server, "message": "no changes"})
		return
	}

	errSave := server.Save("update server options")
	if errSave != nil {
		RespondInterface(w, errSave)
		return
	}

	RespondSuccess(w, map[string]interface{}{"server": server})
}

// SPAToggleController handles toggle operations for server/webhook/rabbitmq
// POST /api/toggle
func SPAToggleController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	// Read body once and reuse
	var jb map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&jb)

	// Get token
	token := r.URL.Query().Get("token")
	if token == "" {
		r.ParseForm()
		token = r.Form.Get("token")
	}
	if token == "" {
		if v, ok := jb["token"].(string); ok {
			token = v
		}
	}
	if token == "" {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondErrorCode(w, gerr, http.StatusNotFound)
		return
	}
	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	// Get key
	key := r.URL.Query().Get("key")
	if key == "" {
		r.ParseForm()
		key = r.Form.Get("key")
	}
	if key == "" {
		if v, ok := jb["key"].(string); ok {
			key = v
		}
	}
	if key == "" {
		RespondErrorCode(w, errors.New("missing key parameter"), http.StatusBadRequest)
		return
	}

	var err2 error
	// If user tries to toggle server-level features via SPA toggle, instruct them to use /command endpoint instead
	if strings.HasPrefix(key, "server-") {
		// mapping to command actions
		mapping := map[string]string{
			"server-groups":        "groups",
			"server-broadcasts":    "broadcasts",
			"server-readreceipts":  "readreceipts",
			"server-calls":         "calls",
			"server-readupdate":    "readupdate",
		}
		if action, ok := mapping[key]; ok {
			resp := &models.QpResponse{}
			resp.ParseError(fmt.Errorf("use /command with action='%s' instead of /toggle key '%s'", action, key))
			RespondInterfaceCode(w, resp, http.StatusBadRequest)
			return
		}
	}
	if strings.HasPrefix(key, "server") {
		// Server-level toggles are deprecated via SPA toggle endpoint. Use /api/command with action=start|stop or other actions.
		RespondErrorCode(w, fmt.Errorf("server-level toggles are removed from /api/toggle; use /api/command with action=start|stop|restart|groups|..."), http.StatusBadRequest)
		return
	} else if strings.HasPrefix(key, "webhook") {
		url := r.URL.Query().Get("url")
		if url == "" {
			r.ParseForm()
			url = r.Form.Get("url")
		}
		if url == "" {
			if v, ok := jb["url"].(string); ok {
				url = v
			}
		}
		if url == "" {
			RespondErrorCode(w, errors.New("missing url parameter"), http.StatusBadRequest)
			return
		}
		dispatching := server.GetDispatching(url)
		var webhook *models.QpWhatsappServerDispatching = nil
		if dispatching != nil && dispatching.IsWebhook() {
			webhook = models.NewQpWhatsappServerDispatchingFromDispatching(dispatching, server)
		}
		if webhook == nil {
			RespondErrorCode(w, errors.New("webhook not found for url: "+url), http.StatusNotFound)
			return
		}
		switch key {
		case "webhook-forwardinternal":
			_, err2 = webhook.ToggleForwardInternal()
		case "webhook-broadcasts":
			err2 = models.ToggleBroadcasts(webhook)
		case "webhook-groups":
			err2 = models.ToggleGroups(webhook)
		case "webhook-readreceipts":
			err2 = models.ToggleReadReceipts(webhook)
		case "webhook-calls":
			err2 = models.ToggleCalls(webhook)
		default:
			err2 = errors.New("invalid webhook key: " + key)
		}
	} else if strings.HasPrefix(key, "rabbitmq") {
		cs := r.URL.Query().Get("connectionString")
		if cs == "" {
			r.ParseForm()
			cs = r.Form.Get("connectionString")
		}
		if cs == "" {
			if v, ok := jb["connectionString"].(string); ok {
				cs = v
			}
			if cs == "" {
				if v2, ok2 := jb["connection_string"].(string); ok2 {
					cs = v2
				}
			}
		}
		if cs == "" {
			RespondErrorCode(w, errors.New("missing connectionString parameter"), http.StatusBadRequest)
			return
		}
		dispatching := server.GetDispatching(cs)
		var rabbitmq *models.QpWhatsappServerDispatching = nil
		if dispatching != nil && dispatching.IsRabbitMQ() {
			rabbitmq = models.NewQpWhatsappServerDispatchingFromDispatching(dispatching, server)
		}
		if rabbitmq == nil {
			RespondErrorCode(w, errors.New("rabbitmq configuration not found for connection: "+cs), http.StatusNotFound)
			return
		}
		switch key {
		case "rabbitmq-forwardinternal":
			_, err2 = rabbitmq.ToggleForwardInternal()
		case "rabbitmq-broadcasts":
			err2 = models.ToggleBroadcasts(rabbitmq)
		case "rabbitmq-groups":
			err2 = models.ToggleGroups(rabbitmq)
		case "rabbitmq-readreceipts":
			err2 = models.ToggleReadReceipts(rabbitmq)
		case "rabbitmq-calls":
			err2 = models.ToggleCalls(rabbitmq)
		default:
			err2 = errors.New("invalid rabbitmq key: " + key)
		}
	} else {
		err2 = errors.New("invalid key or prefix: " + key)
	}

	if err2 != nil {
		RespondErrorCode(w, err2, http.StatusInternalServerError)
		return
	}

	RespondSuccess(w, map[string]interface{}{"result": "success"})
}

// SPAUserController creates a new user
// POST /api/user
func SPAUserController(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	if email == "" || password == "" {
		var j struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&j); err == nil {
			email = j.Email
			password = j.Password
		}
	}

	if email == "" || password == "" {
		RespondErrorCode(w, errors.New("email and password required"), http.StatusBadRequest)
		return
	}

	exists, err := models.WhatsappService.DB.Users.Exists(email)
	if err != nil {
		RespondInterface(w, err)
		return
	}

	if exists {
		RespondErrorCode(w, errors.New("user already exists"), http.StatusConflict)
		return
	}

	_, cerr := models.WhatsappService.DB.Users.Create(email, password)
	if cerr != nil {
		RespondInterface(w, cerr)
		return
	}

	RespondSuccess(w, map[string]interface{}{"result": "success"})
}

// SPAServerSendController sends a message
// POST /api/server/{token}/send
func SPAServerSendController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token := chi.URLParam(r, "token")
	if len(token) == 0 {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondErrorCode(w, gerr, http.StatusNotFound)
		return
	}

	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	type payload struct {
		Recipient string `json:"recipient"`
		Message   string `json:"message"`
		Id        string `json:"id,omitempty"`
	}

	var p payload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	if p.Recipient == "" || p.Message == "" {
		RespondErrorCode(w, errors.New("recipient and message are required"), http.StatusBadRequest)
		return
	}

	msg, merr := models.ToWhatsappMessage(p.Recipient, p.Message, nil)
	if merr != nil {
		RespondServerError(server, w, merr)
		return
	}

	msg.Id = p.Id
	_, sendErr := server.SendMessage(msg)
	if sendErr != nil {
		RespondServerError(server, w, sendErr)
		return
	}

	RespondSuccess(w, map[string]interface{}{"result": "success", "id": msg.GetId()})
}

// SPAVerifyWebSocketController handles websocket connections for QR code verification
// GET /api/verify/ws
func SPAVerifyWebSocketController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		if err == models.ErrFormUnauthenticated {
			RespondUnauthorized(w, err)
			return
		}
		RespondInterface(w, err)
		return
	}

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("(websocket): service error: %s", err.Error())
		return
	}

	HSDString := r.URL.Query().Get("historysyncdays")
	historysyncdays, _ := strconv.ParseUint(HSDString, 10, 32)

	pairing := &models.QpWhatsappPairing{
		Username:        user.Username,
		HistorySyncDays: uint32(historysyncdays),
	}

	WebSocketStart(*pairing, conn)
}

// SPAUsersListController returns all users created by the current user
// GET /api/users
func SPAUsersListController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	// Build response without exposing password hashes
	type userResponse struct {
		Username  string `json:"username"`
		CreatedBy string `json:"created_by,omitempty"`
		Timestamp string `json:"timestamp,omitempty"`
		IsSelf    bool   `json:"is_self,omitempty"`
	}

	response := []userResponse{{
		Username:  user.Username,
		CreatedBy: "",
		Timestamp: user.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		IsSelf:    true,
	}}

	RespondSuccess(w, map[string]interface{}{"users": response})
}

// SPAUserDeleteController deletes a user
// DELETE /api/user
func SPAUserDeleteController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	type request struct {
		Username string `json:"username"`
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		RespondErrorCode(w, errors.New("username is required"), http.StatusBadRequest)
		return
	}

	// Prevent deleting self
	if req.Username == user.Username {
		RespondErrorCode(w, errors.New("cannot delete yourself"), http.StatusBadRequest)
		return
	}

	RespondErrorCode(w, errors.New("user deletion is not supported in this build"), http.StatusNotImplemented)
}

// SPAEnvironmentController returns all environment variables grouped by category
// GET /api/environment
func SPAEnvironmentController(w http.ResponseWriter, r *http.Request) {
	type envVar struct {
		Name        string `json:"name"`
		Value       string `json:"value"`
		Description string `json:"description"`
	}

	type category struct {
		Name      string   `json:"name"`
		Variables []envVar `json:"variables"`
	}

	categories := []category{
		{
			Name: "Database",
			Variables: []envVar{
				{Name: "DBDRIVER", Value: getEnvMasked("DBDRIVER", false), Description: "Database driver (sqlite3, mysql, postgres)"},
				{Name: "DBHOST", Value: getEnvMasked("DBHOST", false), Description: "Database host"},
				{Name: "DBDATABASE", Value: getEnvMasked("DBDATABASE", false), Description: "Database name"},
				{Name: "DBPORT", Value: getEnvMasked("DBPORT", false), Description: "Database port"},
				{Name: "DBUSER", Value: getEnvMasked("DBUSER", false), Description: "Database user"},
				{Name: "DBPASSWORD", Value: getEnvMasked("DBPASSWORD", true), Description: "Database password"},
				{Name: "DBSSLMODE", Value: getEnvMasked("DBSSLMODE", false), Description: "Database SSL mode"},
			},
		},
		{
			Name: "API",
			Variables: []envVar{
				{Name: "MASTERKEY", Value: getEnvMasked("MASTERKEY", true), Description: "Master key for API authentication"},
				{Name: "API_PREFIX", Value: getEnvMasked("API_PREFIX", false), Description: "API routes prefix (default: api)"},
				{Name: "API_TIMEOUT", Value: getEnvMasked("API_TIMEOUT", false), Description: "API request timeout in milliseconds"},
				{Name: "SIGNING_SECRET", Value: getEnvMasked("SIGNING_SECRET", true), Description: "Token for hash signing cookies"},
				{Name: "WEBSOCKETSSL", Value: getEnvMasked("WEBSOCKETSSL", false), Description: "Use SSL for websocket QR code"},
				{Name: "WEBHOOK_TIMEOUT", Value: getEnvMasked("WEBHOOK_TIMEOUT", false), Description: "Webhook timeout in milliseconds"},
			},
		},
		{
			Name: "WebServer",
			Variables: []envVar{
				{Name: "WEBSERVER_HOST", Value: getEnvMasked("WEBSERVER_HOST", false), Description: "HTTP host address"},
				{Name: "WEBSERVER_PORT", Value: getEnvMasked("WEBSERVER_PORT", false), Description: "HTTP port"},
				{Name: "WEBSERVER_LOGS", Value: getEnvMasked("WEBSERVER_LOGS", false), Description: "Enable HTTP request logging"},
				{Name: "WEBAPIHOST", Value: getEnvMasked("WEBAPIHOST", false), Description: "HTTP host (legacy fallback)"},
				{Name: "WEBAPIPORT", Value: getEnvMasked("WEBAPIPORT", false), Description: "HTTP port (legacy fallback)"},
			},
		},
		{
			Name: "Form",
			Variables: []envVar{
				{Name: "FORM", Value: getEnvMasked("FORM", false), Description: "Enable/disable form interface"},
				{Name: "FORM_PREFIX", Value: getEnvMasked("FORM_PREFIX", false), Description: "Form endpoint path prefix (default: form)"},
				{Name: "SERVERS_VIEW_MODE", Value: getEnvMasked("SERVERS_VIEW_MODE", false), Description: "Servers view mode: card or table"},
			},
		},
		{
			Name: "Swagger",
			Variables: []envVar{
				{Name: "SWAGGER", Value: getEnvMasked("SWAGGER", false), Description: "Swagger documentation enabled"},
				{Name: "SWAGGER_PREFIX", Value: getEnvMasked("SWAGGER_PREFIX", false), Description: "Swagger path prefix (default: swagger)"},
			},
		},
		{
			Name: "Metrics",
			Variables: []envVar{
				{Name: "METRICS", Value: getEnvMasked("METRICS", false), Description: "Prometheus metrics enabled"},
				{Name: "METRICS_PREFIX", Value: getEnvMasked("METRICS_PREFIX", false), Description: "Metrics endpoint path prefix"},
				{Name: "METRICS_DASHBOARD", Value: getEnvMasked("METRICS_DASHBOARD", false), Description: "Dashboard endpoint enabled"},
				{Name: "METRICS_DASHBOARD_PREFIX", Value: getEnvMasked("METRICS_DASHBOARD_PREFIX", false), Description: "Dashboard endpoint path prefix"},
			},
		},
		{
			Name: "WhatsApp",
			Variables: []envVar{
				{Name: "READUPDATE", Value: getEnvMasked("READUPDATE", false), Description: "Mark chat read when sending message"},
				{Name: "READRECEIPTS", Value: getEnvMasked("READRECEIPTS", false), Description: "Dispatch read receipts events"},
				{Name: "CALLS", Value: getEnvMasked("CALLS", false), Description: "Accept/handle calls"},
				{Name: "GROUPS", Value: getEnvMasked("GROUPS", false), Description: "Handle group messages"},
				{Name: "BROADCASTS", Value: getEnvMasked("BROADCASTS", false), Description: "Handle broadcast messages"},
				{Name: "HISTORYSYNCDAYS", Value: getEnvMasked("HISTORYSYNCDAYS", false), Description: "History sync days"},
				{Name: "PRESENCE", Value: getEnvMasked("PRESENCE", false), Description: "Presence status"},
				{Name: "WAKEUP_HOUR", Value: getEnvMasked("WAKEUP_HOUR", false), Description: "Scheduled hour(s) to activate presence"},
				{Name: "WAKEUP_DURATION", Value: getEnvMasked("WAKEUP_DURATION", false), Description: "Duration in seconds for wake up"},
			},
		},
		{
			Name: "Whatsmeow",
			Variables: []envVar{
				{Name: "WHATSMEOW_LOGLEVEL", Value: getEnvMasked("WHATSMEOW_LOGLEVEL", false), Description: "Whatsmeow log level"},
				{Name: "WHATSMEOW_DBLOGLEVEL", Value: getEnvMasked("WHATSMEOW_DBLOGLEVEL", false), Description: "Whatsmeow DB log level"},
				{Name: "DISPATCHUNHANDLED", Value: getEnvMasked("DISPATCHUNHANDLED", false), Description: "Dispatch unhandled messages"},
			},
		},
		{
			Name: "General",
			Variables: []envVar{
				{Name: "MIGRATIONS", Value: getEnvMasked("MIGRATIONS", false), Description: "Enable database migrations (true/false/path)"},
				{Name: "APP_TITLE", Value: getEnvMasked("APP_TITLE", false), Description: "Application title for WhatsApp ID"},
				{Name: "REMOVEDIGIT9", Value: getEnvMasked("REMOVEDIGIT9", false), Description: "Remove 9th digit from Brazilian numbers"},
				{Name: "SYNOPSISLENGTH", Value: getEnvMasked("SYNOPSISLENGTH", false), Description: "Synopsis length for messages"},
				{Name: "CACHELENGTH", Value: getEnvMasked("CACHELENGTH", false), Description: "Cache max items"},
				{Name: "CACHEDAYS", Value: getEnvMasked("CACHEDAYS", false), Description: "Cache max days"},
				{Name: "CONVERT_WAVE_TO_OGG", Value: getEnvMasked("CONVERT_WAVE_TO_OGG", false), Description: "Convert wave to OGG"},
				{Name: "COMPATIBLE_MIME_AS_AUDIO", Value: getEnvMasked("COMPATIBLE_MIME_AS_AUDIO", false), Description: "Treat compatible MIME as audio"},
				{Name: "ACCOUNTSETUP", Value: getEnvMasked("ACCOUNTSETUP", false), Description: "Enable account creation"},
				{Name: "LOGLEVEL", Value: getEnvMasked("LOGLEVEL", false), Description: "General log level"},
			},
		},
		{
			Name: "Login Customization",
			Variables: []envVar{
				{Name: "LOGIN_LOGO", Value: getEnvMasked("LOGIN_LOGO", false), Description: "URL for login page logo"},
				{Name: "LOGIN_SUBTITLE", Value: getEnvMasked("LOGIN_SUBTITLE", false), Description: "Subtitle under logo"},
				{Name: "LOGIN_WARNING", Value: getEnvMasked("LOGIN_WARNING", false), Description: "Prominent warning text"},
				{Name: "LOGIN_FOOTER", Value: getEnvMasked("LOGIN_FOOTER", false), Description: "Footer text"},
				{Name: "LOGIN_LAYOUT", Value: getEnvMasked("LOGIN_LAYOUT", false), Description: "Layout type: center|split|simple"},
				{Name: "LOGIN_CUSTOM_CSS", Value: getEnvMasked("LOGIN_CUSTOM_CSS", false), Description: "URL to custom CSS"},
			},
		},
		{
			Name: "Branding",
			Variables: []envVar{
				{Name: "BRANDING_TITLE", Value: getEnvMasked("BRANDING_TITLE", false), Description: "Application title"},
				{Name: "BRANDING_LOGO", Value: getEnvMasked("BRANDING_LOGO", false), Description: "Logo URL"},
				{Name: "BRANDING_FAVICON", Value: getEnvMasked("BRANDING_FAVICON", false), Description: "Favicon URL"},
				{Name: "BRANDING_PRIMARY_COLOR", Value: getEnvMasked("BRANDING_PRIMARY_COLOR", false), Description: "Primary color"},
				{Name: "BRANDING_SECONDARY_COLOR", Value: getEnvMasked("BRANDING_SECONDARY_COLOR", false), Description: "Secondary color"},
				{Name: "BRANDING_ACCENT_COLOR", Value: getEnvMasked("BRANDING_ACCENT_COLOR", false), Description: "Accent color"},
				{Name: "BRANDING_COMPANY_NAME", Value: getEnvMasked("BRANDING_COMPANY_NAME", false), Description: "Company name for footer"},
				{Name: "BRANDING_COMPANY_URL", Value: getEnvMasked("BRANDING_COMPANY_URL", false), Description: "Company URL for footer"},
			},
		},
		{
			Name: "MCP",
			Variables: []envVar{
				{Name: "MCP_ENABLED", Value: getEnvMasked("MCP_ENABLED", false), Description: "MCP server enabled"},
				{Name: "MCP_PATH", Value: getEnvMasked("MCP_PATH", false), Description: "MCP endpoint path"},
			},
		},
		{
			Name: "RabbitMQ",
			Variables: []envVar{
				{Name: "RABBITMQ_QUEUE", Value: getEnvMasked("RABBITMQ_QUEUE", false), Description: "RabbitMQ queue name"},
				{Name: "RABBITMQ_CONNECTIONSTRING", Value: getEnvMasked("RABBITMQ_CONNECTIONSTRING", true), Description: "RabbitMQ connection string"},
				{Name: "RABBITMQ_CACHELENGTH", Value: getEnvMasked("RABBITMQ_CACHELENGTH", false), Description: "RabbitMQ cache length"},
			},
		},
		{
			Name: "SIP Proxy",
			Variables: []envVar{
				{Name: "SIPPROXY_HOST", Value: getEnvMasked("SIPPROXY_HOST", false), Description: "SIP server host (required for activation)"},
				{Name: "SIPPROXY_PORT", Value: getEnvMasked("SIPPROXY_PORT", false), Description: "SIP server port"},
				{Name: "SIPPROXY_PROTOCOL", Value: getEnvMasked("SIPPROXY_PROTOCOL", false), Description: "SIP server protocol"},
				{Name: "SIPPROXY_LOCALPORT", Value: getEnvMasked("SIPPROXY_LOCALPORT", false), Description: "Local SIP port"},
				{Name: "SIPPROXY_PUBLICIP", Value: getEnvMasked("SIPPROXY_PUBLICIP", false), Description: "Public IP for SIP"},
				{Name: "SIPPROXY_STUNSERVER", Value: getEnvMasked("SIPPROXY_STUNSERVER", false), Description: "STUN server for NAT discovery"},
				{Name: "SIPPROXY_MEDIAPORTS", Value: getEnvMasked("SIPPROXY_MEDIAPORTS", false), Description: "RTP media port range"},
				{Name: "SIPPROXY_CODECS", Value: getEnvMasked("SIPPROXY_CODECS", false), Description: "Supported audio codecs"},
				{Name: "SIPPROXY_LOGLEVEL", Value: getEnvMasked("SIPPROXY_LOGLEVEL", false), Description: "SIP proxy log level"},
			},
		},
	}

	RespondSuccess(w, map[string]interface{}{"categories": categories})
}

// getEnvMasked returns the environment variable value, masking sensitive data
func getEnvMasked(name string, sensitive bool) string {
	value := os.Getenv(name)
	if sensitive && value != "" {
		return "********"
	}
	return value
}

// SPAServerMessagesController returns messages for a server (SPA endpoint)
//
//	@Summary		Get server messages (SPA)
//	@Description	Retrieves messages from WhatsApp server with pagination, timestamp and exceptions filtering (authenticated SPA endpoint)
//	@Tags			SPA
//	@Accept			json
//	@Produce		json
//	@Param			token		path		string	true	"Server token"
//	@Param			timestamp	query		string	false	"Timestamp filter for messages (Unix timestamp)"
//	@Param			exceptions	query		string	false	"Filter by exceptions: 'true' for messages with errors, 'false' for messages without errors, omit for all"
//	@Param			page		query		int		false	"Page number (default: 1)"
//	@Param			limit		query		int		false	"Messages per page (default: 50, max: 500)"
//	@Success		200			{object}	models.QpReceiveResponse
//	@Failure		400			{object}	models.QpResponse
//	@Failure		401			{object}	models.QpResponse
//	@Failure		403			{object}	models.QpResponse
//	@Failure		503			{object}	models.QpResponse
//	@Security		Bearer
//	@Router			/api/server/{token}/messages [get]
func SPAServerMessagesController(w http.ResponseWriter, r *http.Request) {
	response := &models.QpReceiveResponse{}

	// Authenticate user via JWT
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		log.Debugf("SPA messages: authentication failed - %v", err)
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	// Get server token from URL path
	token := chi.URLParam(r, "token")
	if len(token) == 0 {
		log.Warnf("SPA messages: missing token parameter")
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	// Find server by token
	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		log.Warnf("SPA messages: server not found for token %s", token)
		RespondInterface(w, gerr)
		return
	}

	// Verify ownership
	if server.User != user.Username {
		log.Warnf("SPA messages: user %s attempted to access server owned by %s", user.Username, server.User)
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	// Get current status
	status := server.GetStatus()
	log.Debugf("SPA messages: server %s status=%s, user=%s", token, status.String(), user.Username)

	// Check if server is ready
	if status != whatsapp.Ready {
		err = fmt.Errorf("server (%s) not ready yet ! current status: %s", server.Wid, status.String())
		log.Infof("SPA messages: server not ready - %s", err.Error())
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
		return
	}

	// Verify handler is attached
	if server.Handler == nil {
		err = fmt.Errorf("handlers not attached")
		log.Errorf("SPA messages: server %s has no handler attached", token)
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Get pagination parameters
	queryValues := r.URL.Query()
	page := 1
	limit := 50 // Default: 50 messages per page

	if pageStr := queryValues.Get("page"); pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	if limitStr := queryValues.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			// Max limit: 500 messages per page
			if limit > 500 {
				limit = 500
			}
		}
	}

	// Get timestamp filter
	timestamp, err := GetTimestamp(r)
	if err != nil {
		log.Warnf("SPA messages: invalid timestamp parameter - %v", err)
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Get exceptions filter parameter
	exceptionsFilter := queryValues.Get("exceptions")

	// Retrieve and filter messages
	allMessages := GetOrderedMessagesWithExceptionsFilter(server, timestamp, exceptionsFilter)
	totalMessages := len(allMessages)

	// Calculate pagination
	totalPages := (totalMessages + limit - 1) / limit // Ceiling division
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	// Calculate start and end indices for current page
	startIdx := (page - 1) * limit
	endIdx := startIdx + limit
	if endIdx > totalMessages {
		endIdx = totalMessages
	}

	// Get messages for current page
	var pagedMessages []whatsapp.WhatsappMessage
	if startIdx < totalMessages {
		pagedMessages = allMessages[startIdx:endIdx]
	} else {
		pagedMessages = []whatsapp.WhatsappMessage{}
	}

	response.Server = server.QpServer
	response.Messages = pagedMessages
	response.Total = uint64(totalMessages)
	response.Page = page
	response.Limit = limit
	response.TotalPages = totalPages

	// Build success message with filter information
	var msg string
	if timestamp > 0 {
		searchTime := time.Unix(timestamp, 0)
		msg = fmt.Sprintf("getting with timestamp: %v => %s", timestamp, searchTime)
	} else {
		msg = "getting without timestamp filter"
	}

	if exceptionsFilter != "" {
		msg += fmt.Sprintf(", exceptions filter: %s", exceptionsFilter)
	}

	msg += fmt.Sprintf(", page %d/%d (%d messages)", page, totalPages, len(pagedMessages))

	log.Debugf("SPA messages: returning %d messages (page %d/%d) for server %s (total: %d)", 
		len(pagedMessages), page, totalPages, token, totalMessages)

	response.ParseSuccess(msg)
	RespondSuccess(w, response)
}

// SPAServerEditMessageController edits a message
// PUT /api/server/{token}/message/{messageid}/edit
func SPAServerEditMessageController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token := chi.URLParam(r, "token")
	if len(token) == 0 {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	messageId := chi.URLParam(r, "messageid")
	if len(messageId) == 0 {
		RespondErrorCode(w, errors.New("missing messageid parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondInterface(w, gerr)
		return
	}

	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	type payload struct {
		Content string `json:"content"`
	}

	var p payload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	if p.Content == "" {
		RespondErrorCode(w, errors.New("content is required"), http.StatusBadRequest)
		return
	}

	if err := server.Edit(messageId, p.Content); err != nil {
		RespondServerError(server, w, err)
		return
	}

	RespondSuccess(w, map[string]interface{}{"result": "success"})
}

// SPAServerRevokeMessageController revokes/deletes a message
// DELETE /api/server/{token}/message/{messageid}
func SPAServerRevokeMessageController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token := chi.URLParam(r, "token")
	if len(token) == 0 {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	messageId := chi.URLParam(r, "messageid")
	if len(messageId) == 0 {
		RespondErrorCode(w, errors.New("missing messageid parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondInterface(w, gerr)
		return
	}

	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	if err := server.Revoke(messageId); err != nil {
		RespondServerError(server, w, err)
		return
	}

	RespondSuccess(w, map[string]interface{}{"result": "success"})
}

// SPAServerArchiveChatController archives a chat
// POST /api/server/{token}/chat/archive
func SPAServerArchiveChatController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token := chi.URLParam(r, "token")
	if len(token) == 0 {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondInterface(w, gerr)
		return
	}

	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	type payload struct {
		ChatId  string `json:"chatid"`
		Archive bool   `json:"archive"`
	}

	var p payload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	if p.ChatId == "" {
		RespondErrorCode(w, errors.New("chatid is required"), http.StatusBadRequest)
		return
	}

	// Format and validate the chat ID
	formattedChatId, fmtErr := whatsapp.FormatEndpoint(p.ChatId)
	if fmtErr != nil {
		RespondErrorCode(w, fmtErr, http.StatusBadRequest)
		return
	}

	// Get WhatsmeowConnection
	conn := server.GetConnection().(*whatsmeow.WhatsmeowConnection)

	// Archive or unarchive chat
	if err := whatsmeow.ArchiveChat(conn, formattedChatId, p.Archive); err != nil {
		RespondServerError(server, w, err)
		return
	}

	RespondSuccess(w, map[string]interface{}{"result": "success"})
}

// SPAServerPresenceController sends presence to a chat
// POST /api/server/{token}/chat/presence
func SPAServerPresenceController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token := chi.URLParam(r, "token")
	if len(token) == 0 {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondInterface(w, gerr)
		return
	}

	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	type payload struct {
		ChatId   string `json:"chatid"`
		Type     string `json:"type"`
		Duration int    `json:"duration"`
	}

	var p payload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	if p.ChatId == "" {
		RespondErrorCode(w, errors.New("chatid is required"), http.StatusBadRequest)
		return
	}

	if p.Type == "" {
		p.Type = "composing"
	}

	// Parse presence type string to enum
	var presenceType whatsapp.WhatsappChatPresenceType
	presenceType.Parse(p.Type)

	if err := server.SendChatPresence(p.ChatId, presenceType); err != nil {
		RespondServerError(server, w, err)
		return
	}

	RespondSuccess(w, map[string]interface{}{"result": "success"})
}

// SPAServerContactsController returns all contacts for a server
// GET /api/server/{token}/contacts
func SPAServerContactsController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		RespondBadRequest(w, fmt.Errorf("missing token parameter"))
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondNotFound(w, fmt.Errorf("server token not found"))
		return
	}

	// Check ownership
	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	// Get contacts
	contacts, err := server.GetContacts()
	if err != nil {
		RespondServerError(server, w, err)
		return
	}

	response := &models.QpContactsResponse{}
	response.Total = len(contacts)
	response.Contacts = contacts
	RespondSuccess(w, response)
}

// SPAServerGroupsController returns all groups for a server
// GET /api/server/{token}/groups
func SPAServerGroupsController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		RespondBadRequest(w, fmt.Errorf("missing token parameter"))
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondNotFound(w, fmt.Errorf("server token not found"))
		return
	}

	// Check ownership
	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	// Get all joined groups
	groups, err := server.GetGroupManager().GetJoinedGroups()
	if err != nil {
		RespondServerError(server, w, err)
		return
	}

	response := &models.QpGroupsResponse{}
	response.Total = len(groups)
	response.Groups = groups
	RespondSuccess(w, response)
}

// SPAServerDownloadMediaController downloads media for a message
// GET /api/server/{token}/download/{messageid}
func SPAServerDownloadMediaController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUserFromRequest(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token := chi.URLParam(r, "token")
	if len(token) == 0 {
		RespondErrorCode(w, errors.New("missing token parameter"), http.StatusBadRequest)
		return
	}

	messageId := chi.URLParam(r, "messageid")
	if len(messageId) == 0 {
		RespondErrorCode(w, errors.New("missing messageid parameter"), http.StatusBadRequest)
		return
	}

	server, gerr := models.GetServerFromToken(token)
	if gerr != nil {
		RespondInterface(w, gerr)
		return
	}

	if server.User != user.Username {
		RespondErrorCode(w, errors.New("server token not owned by user"), http.StatusForbidden)
		return
	}

	// Use existing download logic from server - Download returns attachment with metadata
	att, err := server.Download(messageId, false)
	if err != nil {
		RespondServerError(server, w, err)
		return
	}

	// Set appropriate content type and disposition
	w.Header().Set("Content-Type", att.Mimetype)
	if att.FileName != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", att.FileName))
	}
	content := att.GetContent()
	if content != nil {
		w.Write(*content)
	}
}
