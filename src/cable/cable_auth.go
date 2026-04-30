package cable

import (
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/jwtauth"
	models "github.com/nocodeleaks/quepasa/models"
	runtime "github.com/nocodeleaks/quepasa/runtime"
)

// cableTokenAuth intentionally reuses the same signing secret and token lookup
// behavior already used by the form and SPA HTTP layers. That keeps websocket
// authentication aligned with the browser session cookie named "jwt".
var cableTokenAuth = jwtauth.New("HS256", []byte(os.Getenv(models.ENV_SIGNING_SECRET)), nil)

// CableVerifier exposes the JWT verifier middleware so the route wiring stays
// small and the auth policy remains centralized in this package.
func CableVerifier() func(http.Handler) http.Handler {
	return jwtauth.Verifier(cableTokenAuth)
}

// CableAuthenticator ensures the request already carries a valid JWT before the
// websocket upgrade starts. Rejecting here keeps unauthorized requests in the
// normal HTTP path rather than after the protocol switch.
func CableAuthenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())
		if err != nil || token == nil || !token.Valid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if _, err := GetCableUser(r); err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetCableUser resolves the authenticated user from the JWT claims already
// attached to the request context by jwtauth.Verifier.
func GetCableUser(r *http.Request) (*models.QpUser, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return nil, err
	}

	username, ok := claims["user_id"].(string)
	if !ok || strings.TrimSpace(username) == "" {
		return nil, models.ErrFormUnauthenticated
	}

	return runtime.FindPersistedUser(username)
}
