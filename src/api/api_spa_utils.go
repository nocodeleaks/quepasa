package api

import (
	"net/http"

	"github.com/go-chi/jwtauth"
	models "github.com/nocodeleaks/quepasa/models"
)

// GetFormUserFromRequest gets the user_id from the JWT and finds the
// corresponding user in the database. This is used by SPA controllers
// that require JWT authentication.
func GetFormUserFromRequest(r *http.Request) (*models.QpUser, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return nil, err
	}

	user, ok := claims["user_id"].(string)
	if !ok {
		return nil, models.ErrFormUnauthenticated
	}

	return models.WhatsappService.DB.Users.Find(user)
}
