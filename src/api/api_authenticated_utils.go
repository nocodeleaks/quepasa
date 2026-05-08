package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	models "github.com/nocodeleaks/quepasa/models"
	runtime "github.com/nocodeleaks/quepasa/runtime"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

type authenticatedAPIContextKey string

const scopedSessionAuthKey authenticatedAPIContextKey = "authenticated_scoped_session_auth"

type scopedSessionAuth struct {
	Token    string
	Username string
}

func withScopedSessionAuth(r *http.Request, token string, username string) *http.Request {
	auth := scopedSessionAuth{
		Token:    strings.TrimSpace(token),
		Username: strings.TrimSpace(username),
	}
	ctx := context.WithValue(r.Context(), scopedSessionAuthKey, auth)
	return r.WithContext(ctx)
}

func getScopedSessionAuth(r *http.Request) (scopedSessionAuth, bool) {
	if r == nil {
		return scopedSessionAuth{}, false
	}

	raw := r.Context().Value(scopedSessionAuthKey)
	auth, ok := raw.(scopedSessionAuth)
	if !ok {
		return scopedSessionAuth{}, false
	}

	auth.Token = strings.TrimSpace(auth.Token)
	auth.Username = strings.TrimSpace(auth.Username)
	if auth.Token == "" || auth.Username == "" {
		return scopedSessionAuth{}, false
	}

	return auth, true
}

func getScopedSessionToken(r *http.Request) (string, bool) {
	auth, ok := getScopedSessionAuth(r)
	if !ok {
		return "", false
	}
	return auth.Token, true
}

func ensureTokenScope(r *http.Request, token string) error {
	scopedToken, scoped := getScopedSessionToken(r)
	if !scoped {
		return nil
	}

	resolved := strings.TrimSpace(token)
	if resolved == "" {
		return errors.New("missing token parameter")
	}

	if !strings.EqualFold(resolved, scopedToken) {
		return errors.New("server token not owned by user")
	}

	return nil
}

// GetAuthenticatedUser resolves the authenticated user for authenticated API routes from the JWT claims.
// This duplicates the minimum auth lookup logic from the form package so the API
// layer can stay independent and avoid a package cycle with form.
func GetAuthenticatedUser(r *http.Request) (*models.QpUser, error) {
	if scopedAuth, ok := getScopedSessionAuth(r); ok {
		return findPersistedUser(scopedAuth.Username)
	}

	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return nil, err
	}

	username, ok := claims["user_id"].(string)
	if !ok || strings.TrimSpace(username) == "" {
		return nil, models.ErrFormUnauthenticated
	}

	return findPersistedUser(username)
}

// GetAuthenticatedTokenParam returns the server token from an authenticated route and validates presence.
func GetAuthenticatedTokenParam(r *http.Request) (string, error) {
	scopedAuth, hasScopedAuth := getScopedSessionAuth(r)
	token := strings.TrimSpace(chi.URLParam(r, "token"))
	if hasScopedAuth {
		return scopedAuth.Token, nil
	}
	if token == "" {
		return "", errors.New("missing token parameter")
	}

	return token, nil
}

// GetOwnedServerRecord returns the persisted server record for a token and ensures
// the authenticated user is allowed to access it.
func GetOwnedServerRecord(user *models.QpUser, token string) (*models.QpServer, error) {
	resolvedToken := strings.TrimSpace(token)
	if decodedToken, decodeErr := url.PathUnescape(resolvedToken); decodeErr == nil {
		decodedToken = strings.TrimSpace(decodedToken)
		if decodedToken != "" {
			resolvedToken = decodedToken
		}
	}

	server, err := findPersistedServerRecord(resolvedToken)
	if err != nil {
		return nil, err
	}

	if server.GetUser() != user.Username {
		return nil, errors.New("server token not owned by user")
	}

	return server, nil
}

// FindLiveServer returns the in-memory live server instance when present. Missing
// live state is not treated as an error because some authenticated reads must still work for
// disconnected servers that only exist in the database.
func FindLiveServer(token string) *models.QpWhatsappServer {
	server, ok := runtime.FindLiveSessionByToken(token)
	if !ok {
		return nil
	}
	return server
}

// GetOwnedLiveServer returns the in-memory server instance only after the user
// has been authorized against the persisted server record.
func GetOwnedLiveServer(user *models.QpUser, token string) (*models.QpWhatsappServer, error) {
	if _, err := GetOwnedServerRecord(user, token); err != nil {
		return nil, err
	}

	server := FindLiveServer(token)
	if server == nil {
		return nil, fmt.Errorf("server is not active in memory")
	}

	return server, nil
}

// EnsureLiveServerReady validates that the live server can serve realtime/message operations.
func EnsureLiveServerReady(server *models.QpWhatsappServer) error {
	if server == nil {
		return fmt.Errorf("server is not active in memory")
	}

	if server.GetStatus() != whatsapp.Ready {
		return &ApiServerNotReadyException{Wid: server.GetWId(), Status: server.GetStatus()}
	}

	if server.Handler == nil {
		return fmt.Errorf("handlers not attached")
	}

	if _, err := server.GetValidConnection(); err != nil {
		return err
	}

	return nil
}

// CountDispatchingForServer counts dispatching rows from the current live server
// when available, otherwise it falls back to persisted dispatching data.
func CountDispatchingForServer(token string, liveServer *models.QpWhatsappServer) (dispatchCount, webhookCount, rabbitmqCount int) {
	if liveServer != nil {
		dispatchings := liveServer.GetDispatchingByFilter("")
		dispatchCount = len(dispatchings)
		webhookCount = len(liveServer.GetWebhooks())
		rabbitmqCount = len(liveServer.GetRabbitMQConfigsByQueue(""))
		return
	}

	dispatchings, err := runtime.FindPersistedDispatching(token)
	if err != nil {
		return
	}

	dispatchCount = len(dispatchings)
	for _, dispatching := range dispatchings {
		if dispatching == nil || dispatching.QpDispatching == nil {
			continue
		}

		switch dispatching.Type {
		case models.DispatchingTypeWebhook:
			webhookCount++
		case models.DispatchingTypeRabbitMQ:
			rabbitmqCount++
		}
	}

	return
}

type serverRuntimeSnapshot struct {
	state         whatsapp.WhatsappConnectionState
	timestamps    models.QpTimestamps
	dispatchCount int
	webhookCount  int
	rabbitmqCount int
}

func recoverAPIValue[T any](fallback T, operation string, fields log.Fields, fn func() T) (result T) {
	defer func() {
		if recovered := recover(); recovered != nil {
			entry := log.WithFields(fields)
			entry.WithField("panic", recovered).Errorf("recovered from panic during %s", operation)
			entry.Debugf("panic stack:\n%s", debug.Stack())
			result = fallback
		}
	}()

	return fn()
}

func buildFallbackServerSummary(dbServer *models.QpServer, snap serverRuntimeSnapshot) map[string]interface{} {
	if dbServer == nil {
		return map[string]interface{}{
			"token":          "",
			"wid":            "",
			"state":          snap.state.String(),
			"stateCode":      snap.state.EnumIndex(),
			"verified":       false,
			"devel":          false,
			"user":           "",
			"timestamp":      time.Time{},
			"startTime":      snap.timestamps.Start,
			"lastUpdate":     snap.timestamps.Update,
			"uptimeSeconds":  int64(0),
			"dispatchCount":  snap.dispatchCount,
			"webhookCount":   snap.webhookCount,
			"rabbitmqCount":  snap.rabbitmqCount,
			"hasDispatching": snap.dispatchCount > 0,
			"hasWebhooks":    snap.webhookCount > 0,
			"hasRabbitMQ":    snap.rabbitmqCount > 0,
			"groups":         false,
			"broadcasts":     false,
			"readReceipts":   false,
			"calls":          false,
		}
	}

	uptimeSeconds := int64(0)
	if !snap.timestamps.Start.IsZero() {
		uptimeSeconds = int64(time.Since(snap.timestamps.Start).Seconds())
	}

	return map[string]interface{}{
		"token":          dbServer.Token,
		"wid":            dbServer.GetWId(),
		"state":          snap.state.String(),
		"stateCode":      snap.state.EnumIndex(),
		"verified":       dbServer.Verified,
		"devel":          dbServer.Devel,
		"user":           dbServer.GetUser(),
		"timestamp":      dbServer.Timestamp,
		"startTime":      snap.timestamps.Start,
		"lastUpdate":     snap.timestamps.Update,
		"uptimeSeconds":  uptimeSeconds,
		"dispatchCount":  snap.dispatchCount,
		"webhookCount":   snap.webhookCount,
		"rabbitmqCount":  snap.rabbitmqCount,
		"hasDispatching": snap.dispatchCount > 0,
		"hasWebhooks":    snap.webhookCount > 0,
		"hasRabbitMQ":    snap.rabbitmqCount > 0,
		"groups":         dbServer.GetGroups(),
		"broadcasts":     dbServer.GetBroadcasts(),
		"readReceipts":   dbServer.GetReadReceipts(),
		"calls":          dbServer.GetCalls(),
	}
}

// BuildServerSummary creates a stable JSON-friendly server summary for SPA reads.
func BuildServerSummary(dbServer *models.QpServer, liveServer *models.QpWhatsappServer) map[string]interface{} {
	fallbackRuntime := serverRuntimeSnapshot{
		state: whatsapp.Disconnected,
	}
	if dbServer != nil {
		fallbackRuntime.dispatchCount, fallbackRuntime.webhookCount, fallbackRuntime.rabbitmqCount = CountDispatchingForServer(dbServer.Token, nil)
	}

	fields := log.Fields{}
	if dbServer != nil {
		fields["token"] = dbServer.Token
		fields["wid"] = dbServer.GetWId()
		fields["user"] = dbServer.GetUser()
	}

	snap := recoverAPIValue(fallbackRuntime, "BuildServerSummary runtime snapshot", fields, func() serverRuntimeSnapshot {
		snapshot := fallbackRuntime
		if liveServer != nil {
			snapshot.state = liveServer.GetState()
			snapshot.timestamps = liveServer.Timestamps
		}
		if dbServer != nil {
			snapshot.dispatchCount, snapshot.webhookCount, snapshot.rabbitmqCount = CountDispatchingForServer(dbServer.Token, liveServer)
		}
		return snapshot
	})

	return recoverAPIValue(buildFallbackServerSummary(dbServer, snap), "BuildServerSummary response payload", fields, func() map[string]interface{} {
		return buildFallbackServerSummary(dbServer, snap)
	})
}
