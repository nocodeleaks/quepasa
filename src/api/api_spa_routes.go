package api

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	models "github.com/nocodeleaks/quepasa/models"
)

// Token of authentication / encryption for SPA routes
var spaTokenAuth = jwtauth.New("HS256", []byte(os.Getenv(models.ENV_SIGNING_SECRET)), nil)

// GetSPATokenAuth returns the JWT authentication token for SPA routes
func GetSPATokenAuth() *jwtauth.JWTAuth {
	return spaTokenAuth
}

// SPAAuthenticatorHandler is the JWT authentication middleware for SPA routes
func SPAAuthenticatorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())

		if err != nil {
			RespondErrorCode(w, err, http.StatusUnauthorized)
			return
		}

		if token == nil || !token.Valid {
			RespondErrorCode(w, models.ErrFormUnauthenticated, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RegisterSPAControllers registers SPA-related controllers under /api prefix
// These are authenticated routes for the Vue SPA frontend
func RegisterSPAControllers(r chi.Router) {
	tokenAuth := GetSPATokenAuth()
	r.Use(jwtauth.Verifier(tokenAuth))
	r.Use(SPAAuthenticatorHandler)

	// Session and servers
	r.Get("/session", SPASessionController)
	r.Get("/servers", SPAServersController)
	// Server-side search for SPA
	r.Post("/servers/search", SPAServersSearchController)

	// Account
	r.Get("/account", SPAAccountController)
	r.Get("/account/masterkey", SPAMasterKeyController)

	// Server management
	r.Post("/server/create", SPAServerCreateController)
	r.Post("/server/{token}/update", SPAServerUpdateController)
	r.Get("/server/{token}/info", SPAServerInfoController)
	r.Get("/server/{token}/qrcode", SPAServerQRCodeController)
	r.Get("/server/{token}/paircode", SPAServerPairCodeController)
	r.Post("/server/{token}/send", SPAServerSendController)
	r.Get("/server/{token}/messages", SPAServerMessagesController)
	r.Post("/server/{token}/messages/{messageid}/history/download", SPAServerHistoryDownloadController)
	r.Post("/server/{token}/enable", SPAServerEnableController)
	r.Post("/server/{token}/disable", SPAServerDisableController)
	r.Post("/delete", SPAServerDeleteController)
	r.Post("/debug", SPAServerDebugController)
	r.Post("/toggle", SPAToggleController)

	// Message operations
	r.Put("/server/{token}/message/{messageid}/edit", SPAServerEditMessageController)
	r.Delete("/server/{token}/message/{messageid}", SPAServerRevokeMessageController)
	r.Get("/server/{token}/download/{messageid}", SPAServerDownloadMediaController)

	// Chat operations
	r.Post("/server/{token}/chat/archive", SPAServerArchiveChatController)
	r.Post("/server/{token}/chat/presence", SPAServerPresenceController)

	// Contacts and Groups
	r.Get("/server/{token}/contacts", SPAServerContactsController)
	r.Get("/server/{token}/groups", SPAServerGroupsController)

	// User management
	r.Get("/users", SPAUsersListController)
	r.Post("/user", SPAUserController)
	r.Delete("/user", SPAUserDeleteController)

	// Webhooks
	r.Get("/webhooks", SPAWebHooksController)
	r.Post("/webhooks", SPAWebHooksCreateController)
	r.Put("/webhooks", SPAWebHooksUpdateController)
	r.Delete("/webhooks", SPAWebHooksDeleteController)

	// RabbitMQ
	r.Get("/rabbitmq", SPARabbitMQController)
	r.Post("/rabbitmq", SPARabbitMQCreateController)
	r.Delete("/rabbitmq", SPARabbitMQDeleteController)

	// Environment
	r.Get("/environment", SPAEnvironmentController)

	// WebSocket for QR code verification
	r.HandleFunc("/verify/ws", SPAVerifyWebSocketController)
}
