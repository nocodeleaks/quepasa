package form

import (
	"errors"
	"html/template"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	api "github.com/nocodeleaks/quepasa/api"
	viewmodel "github.com/nocodeleaks/quepasa/apps/form/viewmodel"
	environment "github.com/nocodeleaks/quepasa/environment"
	models "github.com/nocodeleaks/quepasa/models"
	webserver "github.com/nocodeleaks/quepasa/webserver"
	log "github.com/sirupsen/logrus"
)

func buildFormPublicEndpoint(suffix string) string {
	base := strings.TrimRight(FormEndpointPrefix, "/")
	return base + suffix
}

var FormLoginEndpoint string = buildFormPublicEndpoint("/login")
var FormSetupEndpoint string = buildFormPublicEndpoint("/setup")
var FormLogoutEndpoint string = buildFormPublicEndpoint("/logout")
var FormDownloadEndpoint string = buildFormPublicEndpoint("/download")

var CanonicalFormLoginEndpoint string = "/apps/form/login"
var CanonicalFormSetupEndpoint string = "/apps/form/setup"
var CanonicalFormLogoutEndpoint string = "/apps/form/logout"

var LegacyFormLoginEndpoint string = "/login"
var LegacyFormSetupEndpoint string = "/setup"
var LegacyFormLogoutEndpoint string = "/logout"

func RegisterFormControllers(r chi.Router) {

	r.Get(FormEndpointPrefix, IndexHandler)
	r.Get(FormEndpointPrefix+"/", IndexHandler)
	r.Get(CanonicalFormEndpointPrefix, IndexHandler)
	r.Get(CanonicalFormEndpointPrefix+"/", IndexHandler)
	r.Get("/", IndexHandler)
	r.With(jwtauth.Verifier(GetTokenAuth())).Get(FormLoginEndpoint, LoginFormHandler)
	r.Post(FormLoginEndpoint, LoginHandler)
	r.With(jwtauth.Verifier(GetTokenAuth())).Get(CanonicalFormLoginEndpoint, LoginFormHandler)
	r.Post(CanonicalFormLoginEndpoint, LoginHandler)
	r.Get(FormLogoutEndpoint, LogoutHandler)
	r.Get(CanonicalFormLogoutEndpoint, LogoutHandler)
	r.With(jwtauth.Verifier(GetTokenAuth())).Get(LegacyFormLoginEndpoint, LoginFormHandler)
	r.Post(LegacyFormLoginEndpoint, LoginHandler)
	r.Get(LegacyFormLogoutEndpoint, LogoutHandler)

	// disable /setup if environment is false
	if environment.Settings.General.AccountSetup {
		r.Get(FormSetupEndpoint, SetupFormHandler)
		r.Post(FormSetupEndpoint, SetupHandler)
		r.Get(CanonicalFormSetupEndpoint, SetupFormHandler)
		r.Post(CanonicalFormSetupEndpoint, SetupHandler)
		r.Get(LegacyFormSetupEndpoint, SetupFormHandler)
		r.Post(LegacyFormSetupEndpoint, SetupHandler)
	}
}

// LoginFormHandler renders route GET "/apps/form/login"
func LoginFormHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUser(r)
	if err == nil && user != nil {
		http.Redirect(w, r, FormAccountEndpoint, http.StatusFound)
		return
	}

	data := viewmodel.LoginPageData{
		PageTitle: "Login - Quepasa",
		Version:   models.QpVersion,
		Apps:      webserver.DiscoverFrontendApps(),
	}

	templates := template.Must(template.ParseFiles(GetViewPath("layouts/main.tmpl"), GetViewPath("login.tmpl")))
	templates.ExecuteTemplate(w, "main", data)
}

// LoginHandler renders route POST "/apps/form/login"
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("email")
	password := r.Form.Get("password")

	if username == "" || password == "" {
		api.RespondUnauthorized(w, errors.New("missing username or password"))
		return
	}

	user, err := models.WhatsappService.GetUser(username, password)
	if err != nil {
		api.RespondUnauthorized(w, err)
		return
	}

	claims := jwt.MapClaims{"user_id": user.Username}
	jwtauth.SetIssuedNow(claims)
	jwtauth.SetExpiryIn(claims, 24*time.Hour)

	tokenAuth := GetTokenAuth()
	_, tokenString, err := tokenAuth.Encode(claims)
	if err != nil {
		api.RespondErrorCode(w, errors.New("cannot encode token to save"), 500)
		return
	}

	cookie := &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		MaxAge:   60 * 60 * 24,
		Path:     "/",
		HttpOnly: true,
	}

	log.Debugf("setting cookie and redirecting to: %v", FormAccountEndpoint)
	http.SetCookie(w, cookie)
	http.Redirect(w, r, FormAccountEndpoint, http.StatusFound)
}
