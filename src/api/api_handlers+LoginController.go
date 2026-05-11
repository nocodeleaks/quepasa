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
//
// The current implementation intentionally keeps most branding fields empty because
// we have not imported PR #39 branding/environment customizations yet. The endpoint
// still exists now so a web client can discover basic runtime facts without scraping
// templates or depending on private settings.
func LoginConfigController(w http.ResponseWriter, r *http.Request) {
	appTitle := environment.Settings.General.AppTitle
	if appTitle == "" {
		appTitle = "QuePasa"
	}

	response := map[string]interface{}{
		// Stable fields already supported by the current environment model.
		"appTitle":     appTitle,
		"accountSetup": environment.Settings.General.AccountSetup,
		"version":      models.QpVersion,
		// Reserved compatibility fields expected by the web client branch. They remain empty
		// until we decide to import branding/login customization support explicitly.
		"loginLogo":     "",
		"loginSubtitle": "",
		"loginWarning":  "",
		"loginFooter":   "",
		"loginLayout":   "center",
		"customCss":     "",
		"fontAwesome":   "",
		"googleFonts":   "",
		// SQLite migration discovery from PR #39 is not wired yet, so expose the
		// field with a conservative false value to keep the payload shape stable.
		"sqliteMigrationAvailable": false,
		"branding": map[string]interface{}{
			"title":          appTitle,
			"logo":           "",
			"favicon":        "",
			"primaryColor":   "",
			"secondaryColor": "",
			"accentColor":    "",
			"companyName":    "",
			"companyUrl":     "",
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
