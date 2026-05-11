package api

import (
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	events "github.com/nocodeleaks/quepasa/events"
	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	runtime "github.com/nocodeleaks/quepasa/runtime"
)

// authenticatedTokenAuth reuses the same signing secret as the form login flow so a browser
// session authenticated through the existing UI can call authenticated API endpoints without a
// second token system.
var authenticatedTokenAuth = jwtauth.New("HS256", []byte(os.Getenv(models.ENV_SIGNING_SECRET)), nil)

// GetAuthenticatedTokenAuth returns the JWT authentication token used by authenticated API routes.
func GetAuthenticatedTokenAuth() *jwtauth.JWTAuth {
	return authenticatedTokenAuth
}

// RegisterAuthenticatedPublicControllers exposes authenticated API bootstrap routes that must stay
// reachable before authentication, such as the login screen configuration.
func RegisterAuthenticatedPublicControllers(r chi.Router) {
	r.Get("/login/config", LoginConfigController)
	r.Post("/users", PublicUserCreateController)
}

// AuthenticatedAPIHandler validates JWT-based authenticated API requests after the jwtauth verifier
// has extracted the token from the request context.
func AuthenticatedAPIHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())
		if err == nil && token != nil && token.Valid {
			publishAuthenticatedAPIEvent(r, "success", "validated_jwt")
			next.ServeHTTP(w, r)
			return
		}

		scopedToken := strings.TrimSpace(r.Header.Get(library.HeaderToken))
		if scopedToken == "" {
			reason := "invalid_token"
			if err != nil {
				reason = "jwt_context_error"
			}
			publishAuthenticatedAPIEvent(r, "unauthorized", reason)
			RespondErrorCode(w, models.ErrFormUnauthenticated, http.StatusUnauthorized)
			return
		}

		username := ""
		if server, lookupErr := findPersistedServerRecord(scopedToken); lookupErr == nil && server != nil {
			username = strings.TrimSpace(server.GetUser())
		}
		if username == "" {
			if liveSession, ok := runtime.FindLiveSessionByToken(scopedToken); ok && liveSession != nil {
				username = strings.TrimSpace(liveSession.GetUser())
			}
		}

		if username == "" {
			publishAuthenticatedAPIEvent(r, "unauthorized", "invalid_scoped_session_token")
			RespondErrorCode(w, models.ErrFormUnauthenticated, http.StatusUnauthorized)
			return
		}

		r = withScopedSessionAuth(r, scopedToken, username)
		publishAuthenticatedAPIEvent(r, "success", "validated_session_token")

		next.ServeHTTP(w, r)
	})
}

func publishAuthenticatedAPIEvent(r *http.Request, status string, reason string) {
	events.Publish(events.Event{
		Name:   "api.authenticated.authentication",
		Source: "api.authenticated_authenticator",
		Status: status,
		Attributes: map[string]string{
			"reason": reason,
			"route":  resolveRoutePattern(r),
		},
	})
}

// RegisterAuthenticatedControllers registers the initial authenticated API endpoints.
//
// This is intentionally a narrow slice. We only expose handlers already adapted to
// the current develop branch instead of importing the full legacy browser-specific surface.
func RegisterAuthenticatedControllers(r chi.Router) {
	tokenAuth := GetAuthenticatedTokenAuth()
	r.Use(jwtauth.Verifier(tokenAuth))
	r.Use(AuthenticatedAPIHandler)

	// First extracted authenticated read endpoints.
	r.Get("/session", AuthenticatedSessionController)
	r.Get("/servers", AuthenticatedServersController)
	r.Post("/servers/search", AuthenticatedServersSearchController)
	r.Get("/account", AuthenticatedAccountController)
	r.Get("/account/masterkey", AuthenticatedMasterKeyController)
	r.Post("/master/verify", AuthenticatedMasterVerifyController)
	r.Get("/environment", AuthenticatedEnvironmentController)
	r.Get("/labels", AuthenticatedConversationLabelController)
	r.Post("/labels", AuthenticatedConversationLabelController)
	r.Put("/labels", AuthenticatedConversationLabelController)
	r.Delete("/labels", AuthenticatedConversationLabelController)
	r.Post("/server/create", AuthenticatedServerCreateController)
	r.Get("/server/{token}/info", AuthenticatedServerInfoController)
	r.Get("/server/{token}/qrcode", AuthenticatedServerQRCodeController)
	r.Get("/server/{token}/paircode", AuthenticatedServerPairCodeController)
	r.Patch("/server/{token}", AuthenticatedServerUpdateController)
	r.Delete("/server/{token}", AuthenticatedServerDeleteController)
	r.Post("/server/{token}/debug/toggle", AuthenticatedServerDebugToggleController)
	r.Post("/server/{token}/option/{option}/toggle", AuthenticatedServerOptionToggleController)
	r.Post("/server/{token}/send", AuthenticatedServerSendController)
	r.Get("/server/{token}/chat/labels", AuthenticatedServerConversationLabelController)
	r.Post("/server/{token}/chat/labels", AuthenticatedServerConversationLabelController)
	r.Delete("/server/{token}/chat/labels", AuthenticatedServerConversationLabelController)
	r.Post("/server/{token}/chat/archive", AuthenticatedServerArchiveChatController)
	r.Post("/server/{token}/chat/presence", AuthenticatedServerPresenceController)
	r.Get("/server/{token}/webhooks", AuthenticatedWebHooksController)
	r.Post("/server/{token}/webhooks", AuthenticatedWebHooksController)
	r.Delete("/server/{token}/webhooks", AuthenticatedWebHooksController)
	r.Get("/server/{token}/rabbitmq", AuthenticatedRabbitMQController)
	r.Post("/server/{token}/rabbitmq", AuthenticatedRabbitMQController)
	r.Delete("/server/{token}/rabbitmq", AuthenticatedRabbitMQController)
	r.Get("/users", AuthenticatedUsersListController)
	r.Delete("/user/{username}", AuthenticatedUserDeleteController)
	r.Get("/server/{token}/contacts", AuthenticatedServerContactsController)
	r.Get("/server/{token}/groups", AuthenticatedServerGroupsController)
	r.Get("/server/{token}/picinfo/{chatid}/{pictureid}", AuthenticatedPictureInfoController)
	r.Get("/server/{token}/picinfo/{chatid}", AuthenticatedPictureInfoController)
	r.Post("/server/{token}/groups/create", AuthenticatedGroupCreateController)
	r.Get("/server/{token}/group/{groupid}", AuthenticatedGroupInfoController)
	r.Post("/server/{token}/group/{groupid}/leave", AuthenticatedGroupLeaveController)
	r.Put("/server/{token}/group/{groupid}/name", AuthenticatedGroupNameController)
	r.Put("/server/{token}/group/{groupid}/description", AuthenticatedGroupDescriptionController)
	r.Put("/server/{token}/group/{groupid}/participants", AuthenticatedGroupParticipantsController)
	r.Put("/server/{token}/group/{groupid}/photo", AuthenticatedGroupPhotoController)
	r.With(withCanonicalParams(canonicalGroupIDParam)).Get("/server/{token}/groups/invite", AuthenticatedGroupInviteController)
	r.With(withCanonicalParams(canonicalGroupIDParam)).Delete("/server/{token}/groups/invite", AuthenticatedGroupRevokeInviteController)

	// Current extracted authenticated message/lifecycle/media actions.
	r.Get("/server/{token}/messages", AuthenticatedServerMessagesController)
	r.Put("/server/{token}/message/{messageid}/edit", AuthenticatedServerEditMessageController)
	r.Delete("/server/{token}/message/{messageid}", AuthenticatedServerRevokeMessageController)
	r.Get("/server/{token}/download/{messageid}", AuthenticatedServerDownloadMediaController)
	r.Post("/server/{token}/messages/{messageid}/history/download", AuthenticatedServerHistoryDownloadController)
	r.Post("/server/{token}/enable", AuthenticatedServerEnableController)
	r.Post("/server/{token}/disable", AuthenticatedServerDisableController)
}
