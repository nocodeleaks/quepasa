package api

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
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

// SPAAuthenticatorHandler validates JWT-based SPA requests after the jwtauth verifier
// has extracted the token from the request context.
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
	r.Get("/server/{token}/info", SPAServerInfoController)
	r.Get("/server/{token}/qrcode", SPAServerQRCodeController)
	r.Get("/server/{token}/paircode", SPAServerPairCodeController)
	r.Get("/users", SPAUsersListController)
	r.Get("/server/{token}/contacts", SPAServerContactsController)
	r.Get("/server/{token}/groups", SPAServerGroupsController)

	// First extracted SPA lifecycle/media actions.
	r.Post("/server/{token}/messages/{messageid}/history/download", SPAServerHistoryDownloadController)
	r.Post("/server/{token}/enable", SPAServerEnableController)
	r.Post("/server/{token}/disable", SPAServerDisableController)
}
