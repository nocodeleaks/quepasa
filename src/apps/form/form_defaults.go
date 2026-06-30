package form

import (
	"github.com/go-chi/jwtauth"
	environment "github.com/nocodeleaks/quepasa/environment"
)

// Token of authentication / encryption
var TokenAuth = jwtauth.New("HS256", []byte(environment.Settings.API.SigningSecret), nil)

// GetTokenAuth returns the JWT authentication token
// This is used in the form authenticated controllers
func GetTokenAuth() *jwtauth.JWTAuth {
	return TokenAuth
}
