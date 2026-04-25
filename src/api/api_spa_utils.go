package api

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

// GetSPAUser resolves the authenticated user for SPA-only routes from the JWT claims.
//
// This duplicates the minimum auth lookup logic from the form package so the API
// layer can stay independent and avoid a package cycle with form.
func GetSPAUser(r *http.Request) (*models.QpUser, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return nil, err
	}

	username, ok := claims["user_id"].(string)
	if !ok || strings.TrimSpace(username) == "" {
		return nil, models.ErrFormUnauthenticated
	}

	return models.WhatsappService.DB.Users.Find(username)
}

// GetSPATokenParam returns the server token from a SPA route and validates presence.
func GetSPATokenParam(r *http.Request) (string, error) {
	token := strings.TrimSpace(chi.URLParam(r, "token"))
	if token == "" {
		return "", errors.New("missing token parameter")
	}
	return token, nil
}

// GetSPAOwnedServerRecord returns the persisted server record for a token and ensures
// the authenticated user is allowed to access it.
func GetSPAOwnedServerRecord(user *models.QpUser, token string) (*models.QpServer, error) {
	server, err := models.WhatsappService.DB.Servers.FindByToken(token)
	if err != nil {
		return nil, err
	}

	if server.GetUser() != user.Username {
		return nil, errors.New("server token not owned by user")
	}

	return server, nil
}

// FindSPALiveServer returns the in-memory live server instance when present. Missing
// live state is not treated as an error because some SPA reads must still work for
// disconnected servers that only exist in the database.
func FindSPALiveServer(token string) *models.QpWhatsappServer {
	server, err := models.WhatsappService.FindByToken(token)
	if err != nil {
		return nil
	}
	return server
}

// GetSPAOwnedLiveServer returns the in-memory server instance only after the user
// has been authorized against the persisted server record.
func GetSPAOwnedLiveServer(user *models.QpUser, token string) (*models.QpWhatsappServer, error) {
	if _, err := GetSPAOwnedServerRecord(user, token); err != nil {
		return nil, err
	}

	server := FindSPALiveServer(token)
	if server == nil {
		return nil, fmt.Errorf("server is not active in memory")
	}

	return server, nil
}

// EnsureSPAServerReady validates that the live server can serve realtime/message operations.
func EnsureSPAServerReady(server *models.QpWhatsappServer) error {
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

// CountSPADispatchingForServer counts dispatching rows from the current live server
// when available, otherwise it falls back to persisted dispatching data.
func CountSPADispatchingForServer(token string, liveServer *models.QpWhatsappServer) (dispatchCount, webhookCount, rabbitmqCount int) {
	if liveServer != nil {
		dispatchings := liveServer.GetDispatchingByFilter("")
		dispatchCount = len(dispatchings)
		webhookCount = len(liveServer.GetWebhooks())
		rabbitmqCount = len(liveServer.GetRabbitMQConfigsByQueue(""))
		return
	}

	dispatchings, err := models.WhatsappService.DB.Dispatching.FindAll(token)
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

type spaServerRuntimeSnapshot struct {
	state         whatsapp.WhatsappConnectionState
	timestamps    models.QpTimestamps
	dispatchCount int
	webhookCount  int
	rabbitmqCount int
}

func recoverSPAValue[T any](fallback T, operation string, fields log.Fields, fn func() T) (result T) {
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

func buildSPAFallbackServerSummary(dbServer *models.QpServer, runtime spaServerRuntimeSnapshot) map[string]interface{} {
	if dbServer == nil {
		return map[string]interface{}{
			"token":          "",
			"wid":            "",
			"state":          runtime.state.String(),
			"stateCode":      runtime.state.EnumIndex(),
			"verified":       false,
			"devel":          false,
			"user":           "",
			"timestamp":      time.Time{},
			"startTime":      runtime.timestamps.Start,
			"lastUpdate":     runtime.timestamps.Update,
			"uptimeSeconds":  int64(0),
			"dispatchCount":  runtime.dispatchCount,
			"webhookCount":   runtime.webhookCount,
			"rabbitmqCount":  runtime.rabbitmqCount,
			"hasDispatching": runtime.dispatchCount > 0,
			"hasWebhooks":    runtime.webhookCount > 0,
			"hasRabbitMQ":    runtime.rabbitmqCount > 0,
			"groups":         false,
			"broadcasts":     false,
			"readReceipts":   false,
			"calls":          false,
		}
	}

	uptimeSeconds := int64(0)
	if !runtime.timestamps.Start.IsZero() {
		uptimeSeconds = int64(time.Since(runtime.timestamps.Start).Seconds())
	}

	return map[string]interface{}{
		"token":          dbServer.Token,
		"wid":            dbServer.Wid,
		"state":          runtime.state.String(),
		"stateCode":      runtime.state.EnumIndex(),
		"verified":       dbServer.Verified,
		"devel":          dbServer.Devel,
		"user":           dbServer.User,
		"timestamp":      dbServer.Timestamp,
		"startTime":      runtime.timestamps.Start,
		"lastUpdate":     runtime.timestamps.Update,
		"uptimeSeconds":  uptimeSeconds,
		"dispatchCount":  runtime.dispatchCount,
		"webhookCount":   runtime.webhookCount,
		"rabbitmqCount":  runtime.rabbitmqCount,
		"hasDispatching": runtime.dispatchCount > 0,
		"hasWebhooks":    runtime.webhookCount > 0,
		"hasRabbitMQ":    runtime.rabbitmqCount > 0,
		"groups":         dbServer.GetGroups(),
		"broadcasts":     dbServer.GetBroadcasts(),
		"readReceipts":   dbServer.GetReadReceipts(),
		"calls":          dbServer.GetCalls(),
	}
}

// BuildSPAServerSummary creates a stable JSON-friendly server summary for SPA reads.
func BuildSPAServerSummary(dbServer *models.QpServer, liveServer *models.QpWhatsappServer) map[string]interface{} {
	fallbackRuntime := spaServerRuntimeSnapshot{
		state: whatsapp.Disconnected,
	}
	if dbServer != nil {
		fallbackRuntime.dispatchCount, fallbackRuntime.webhookCount, fallbackRuntime.rabbitmqCount = CountSPADispatchingForServer(dbServer.Token, nil)
	}

	fields := log.Fields{}
	if dbServer != nil {
		fields["token"] = dbServer.Token
		fields["wid"] = dbServer.Wid
		fields["user"] = dbServer.User
	}

	runtime := recoverSPAValue(fallbackRuntime, "BuildSPAServerSummary runtime snapshot", fields, func() spaServerRuntimeSnapshot {
		snapshot := fallbackRuntime
		if liveServer != nil {
			snapshot.state = liveServer.GetState()
			snapshot.timestamps = liveServer.Timestamps
		}
		if dbServer != nil {
			snapshot.dispatchCount, snapshot.webhookCount, snapshot.rabbitmqCount = CountSPADispatchingForServer(dbServer.Token, liveServer)
		}
		return snapshot
	})

	return recoverSPAValue(buildSPAFallbackServerSummary(dbServer, runtime), "BuildSPAServerSummary response payload", fields, func() map[string]interface{} {
		return buildSPAFallbackServerSummary(dbServer, runtime)
	})
}
