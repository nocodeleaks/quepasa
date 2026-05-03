package api

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	events "github.com/nocodeleaks/quepasa/events"
	models "github.com/nocodeleaks/quepasa/models"
)

// spaTokenAuth reuses the same signing secret as the form login flow so a browser
// session authenticated through the existing UI can call SPA endpoints without a
// second token system.
var spaTokenAuth = jwtauth.New("HS256", []byte(os.Getenv(models.ENV_SIGNING_SECRET)), nil)

// GetSPATokenAuth returns the JWT authentication token used by SPA routes.
func GetSPATokenAuth() *jwtauth.JWTAuth {
	return spaTokenAuth
}

// RegisterSPAPublicControllers exposes SPA bootstrap routes that must stay
// reachable before authentication, such as the login screen configuration.
func RegisterSPAPublicControllers(r chi.Router) {
	r.Get("/login/config", LoginConfigController)
	r.Post("/users", SPAPublicUserCreateController)
}

// SPAAuthenticatorHandler validates JWT-based SPA requests after the jwtauth verifier
// has extracted the token from the request context.
func SPAAuthenticatorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())
		if err != nil {
			publishSPAAuthenticationEvent(r, "unauthorized", "jwt_context_error")
			RespondErrorCode(w, err, http.StatusUnauthorized)
			return
		}

		if token == nil || !token.Valid {
			publishSPAAuthenticationEvent(r, "unauthorized", "invalid_token")
			RespondErrorCode(w, models.ErrFormUnauthenticated, http.StatusUnauthorized)
			return
		}

		publishSPAAuthenticationEvent(r, "success", "validated")

		next.ServeHTTP(w, r)
	})
}

func publishSPAAuthenticationEvent(r *http.Request, status string, reason string) {
	events.Publish(events.Event{
		Name:   "api.spa.authentication",
		Source: "api.spa_authenticator",
		Status: status,
		Attributes: map[string]string{
			"reason": reason,
			"route":  resolveRoutePattern(r),
		},
	})
}

// RegisterSPAControllers registers the initial authenticated SPA-only endpoints.
//
// This is intentionally a narrow slice. We only expose handlers already adapted to
// the current develop branch instead of importing the full PR #39 SPA controller set.
func RegisterSPAControllers(r chi.Router) {
	tokenAuth := GetSPATokenAuth()
	r.Use(jwtauth.Verifier(tokenAuth))
	r.Use(SPAAuthenticatorHandler)

	// First extracted SPA read endpoints.
	r.Get("/session", SPASessionController)
	r.Get("/servers", SPAServersController)
	r.Post("/servers/search", SPAServersSearchController)
	r.Get("/account", SPAAccountController)
	r.Get("/account/masterkey", SPAMasterKeyController)
	r.Get("/environment", SPAEnvironmentController)
	r.Get("/labels", SPAConversationLabelController)
	r.Post("/labels", SPAConversationLabelController)
	r.Put("/labels", SPAConversationLabelController)
	r.Delete("/labels", SPAConversationLabelController)
	r.Post("/server/create", SPAServerCreateController)
	r.Get("/server/{token}/info", SPAServerInfoController)
	r.Get("/server/{token}/qrcode", SPAServerQRCodeController)
	r.Get("/server/{token}/paircode", SPAServerPairCodeController)
	r.Patch("/server/{token}", SPAServerUpdateController)
	r.Delete("/server/{token}", SPAServerDeleteController)
	r.Post("/server/{token}/debug/toggle", SPAServerDebugToggleController)
	r.Post("/server/{token}/option/{option}/toggle", SPAServerOptionToggleController)
	r.Post("/server/{token}/send", SPAServerSendController)
	r.Get("/server/{token}/chat/labels", SPAServerConversationLabelController)
	r.Post("/server/{token}/chat/labels", SPAServerConversationLabelController)
	r.Delete("/server/{token}/chat/labels", SPAServerConversationLabelController)
	r.Post("/server/{token}/chat/archive", SPAServerArchiveChatController)
	r.Post("/server/{token}/chat/presence", SPAServerPresenceController)
	r.Get("/server/{token}/webhooks", SPAWebHooksController)
	r.Post("/server/{token}/webhooks", SPAWebHooksController)
	r.Delete("/server/{token}/webhooks", SPAWebHooksController)
	r.Get("/server/{token}/rabbitmq", SPARabbitMQController)
	r.Post("/server/{token}/rabbitmq", SPARabbitMQController)
	r.Delete("/server/{token}/rabbitmq", SPARabbitMQController)
	r.Get("/users", SPAUsersListController)
	r.Delete("/user/{username}", SPAUserDeleteController)
	r.Get("/server/{token}/contacts", SPAServerContactsController)
	r.Get("/server/{token}/groups", SPAServerGroupsController)
	r.Get("/server/{token}/picinfo/{chatid}/{pictureid}", SPAPictureInfoController)
	r.Get("/server/{token}/picinfo/{chatid}", SPAPictureInfoController)
	r.Post("/server/{token}/groups/create", SPAGroupsCreateController)
	r.Get("/server/{token}/group/{groupid}", SPAGroupInfoController)
	r.Post("/server/{token}/group/{groupid}/leave", SPAGroupLeaveController)
	r.Put("/server/{token}/group/{groupid}/name", SPAGroupNameController)
	r.Put("/server/{token}/group/{groupid}/description", SPAGroupDescriptionController)
	r.Put("/server/{token}/group/{groupid}/participants", SPAGroupParticipantsController)
	r.Put("/server/{token}/group/{groupid}/photo", SPAGroupPhotoController)
	r.With(withCanonicalParams(canonicalGroupIDParam)).Get("/server/{token}/groups/invite", SPAGroupInviteController)
	r.With(withCanonicalParams(canonicalGroupIDParam)).Delete("/server/{token}/groups/invite", SPAGroupRevokeInviteController)

	// Current extracted SPA message/lifecycle/media actions.
	r.Get("/server/{token}/messages", SPAServerMessagesController)
	r.Put("/server/{token}/message/{messageid}/edit", SPAServerEditMessageController)
	r.Delete("/server/{token}/message/{messageid}", SPAServerRevokeMessageController)
	r.Get("/server/{token}/download/{messageid}", SPAServerDownloadMediaController)
	r.Post("/server/{token}/messages/{messageid}/history/download", SPAServerHistoryDownloadController)
	r.Post("/server/{token}/enable", SPAServerEnableController)
	r.Post("/server/{token}/disable", SPAServerDisableController)
}
