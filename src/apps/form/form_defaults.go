package form

import (
	"os"

	"github.com/go-chi/jwtauth"
	models "github.com/nocodeleaks/quepasa/models"
)

// Token of authentication / encryption
var TokenAuth = jwtauth.New("HS256", []byte(os.Getenv(models.ENV_SIGNING_SECRET)), nil)

// GetTokenAuth returns the JWT authentication token
// This is used in the form authenticated controllers
func GetTokenAuth() *jwtauth.JWTAuth {
	return TokenAuth
}
