package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
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

	if server.User != user.Username {
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

// BuildSPAServerSummary creates a stable JSON-friendly server summary for SPA reads.
func BuildSPAServerSummary(dbServer *models.QpServer, liveServer *models.QpWhatsappServer) map[string]interface{} {
	state := whatsapp.Disconnected
	var timestamps models.QpTimestamps
	if liveServer != nil {
		state = liveServer.GetState()
		timestamps = liveServer.Timestamps
	}

	dispatchCount, webhookCount, rabbitmqCount := CountSPADispatchingForServer(dbServer.Token, liveServer)

	uptimeSeconds := int64(0)
	if !timestamps.Start.IsZero() {
		uptimeSeconds = int64(time.Since(timestamps.Start).Seconds())
	}

	return map[string]interface{}{
		"token":          dbServer.Token,
		"wid":            dbServer.Wid,
		"state":          state.String(),
		"stateCode":      state.EnumIndex(),
		"verified":       dbServer.Verified,
		"devel":          dbServer.Devel,
		"user":           dbServer.User,
		"timestamp":      dbServer.Timestamp,
		"startTime":      timestamps.Start,
		"lastUpdate":     timestamps.Update,
		"uptimeSeconds":  uptimeSeconds,
		"dispatchCount":  dispatchCount,
		"webhookCount":   webhookCount,
		"rabbitmqCount":  rabbitmqCount,
		"hasDispatching": dispatchCount > 0,
		"hasWebhooks":    webhookCount > 0,
		"hasRabbitMQ":    rabbitmqCount > 0,
		"groups":         dbServer.GetGroups(),
		"broadcasts":     dbServer.GetBroadcasts(),
		"readReceipts":   dbServer.GetReadReceipts(),
		"calls":          dbServer.GetCalls(),
	}
}
