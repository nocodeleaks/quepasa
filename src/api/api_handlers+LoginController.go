package api

import (
	"errors"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
	environment "github.com/nocodeleaks/quepasa/environment"
	models "github.com/nocodeleaks/quepasa/models"
)

// LoginConfigController returns the public bootstrap payload used by web client login screens.
func LoginConfigController(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"appTitle":                 environment.Settings.Branding.Title,
		"accountSetup":             environment.Settings.General.AccountSetup,
		"version":                  models.QpVersion,
		"loginLogo":                environment.Settings.General.LoginLogo,
		"loginSubtitle":            environment.Settings.General.LoginSubtitle,
		"loginWarning":             environment.Settings.General.LoginWarning,
		"loginFooter":              environment.Settings.General.LoginFooter,
		"loginLayout":              environment.Settings.General.LoginLayout,
		"customCss":                environment.Settings.General.LoginCustomCSS,
		"fontAwesome":              environment.Settings.General.LoginFontAwesome,
		"googleFonts":              environment.Settings.General.LoginGoogleFonts,
		"sqliteMigrationAvailable": false,
		"branding": map[string]interface{}{
			"title":          environment.Settings.Branding.Title,
			"logo":           environment.Settings.Branding.Logo,
			"favicon":        environment.Settings.Branding.Favicon,
			"primaryColor":   environment.Settings.Branding.PrimaryColor,
			"secondaryColor": environment.Settings.Branding.SecondaryColor,
			"accentColor":    environment.Settings.Branding.AccentColor,
			"companyName":    environment.Settings.Branding.CompanyName,
			"companyUrl":     environment.Settings.Branding.CompanyUrl,
		},
	}

	RespondSuccess(w, response)
}

// CanonicalLoginPostController authenticates a user and sets the JWT cookie.
// Accepts application/x-www-form-urlencoded or application/json bodies with
// "email" and "password" fields. Returns JSON on success so authenticated API clients do not
// need to follow HTML redirects.
func CanonicalLoginPostController(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("email")
	password := r.FormValue("password")

	if username == "" || password == "" {
		RespondUnauthorized(w, errors.New("missing username or password"))
		return
	}

	user, err := authenticatePersistedUser(username, password)
	if err != nil {
		RespondUnauthorized(w, err)
		return
	}

	claims := jwt.MapClaims{"user_id": user.Username}
	jwtauth.SetIssuedNow(claims)
	jwtauth.SetExpiryIn(claims, 24*time.Hour)

	_, tokenString, err := GetAuthenticatedTokenAuth().Encode(claims)
	if err != nil {
		RespondErrorCode(w, errors.New("cannot encode token"), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		MaxAge:   60 * 60 * 24,
		Path:     "/",
		HttpOnly: true,
	})

	RespondSuccess(w, map[string]interface{}{
		"user": map[string]interface{}{
			"username": user.Username,
			"email":    user.Username,
		},
		"version": models.QpVersion,
	})
}
