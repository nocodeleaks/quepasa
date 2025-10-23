package form

import (
	"errors"
	"html/template"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	api "github.com/nocodeleaks/quepasa/api"
	environment "github.com/nocodeleaks/quepasa/environment"
	models "github.com/nocodeleaks/quepasa/models"
	log "github.com/sirupsen/logrus"
)

var FormLoginEndpoint string = "/login"
var FormSetupEndpoint string = "/setup"
var FormLogoutEndpoint string = "/logout"
var FormDownloadEndpoint string = "/download"

func RegisterFormControllers(r chi.Router) {

	r.Get("/", IndexHandler)
	r.Get(FormLoginEndpoint, LoginFormHandler)
	r.Post(FormLoginEndpoint, LoginHandler)
	r.Get(FormLogoutEndpoint, LogoutHandler)

	// disable /setup if environment is false
	if environment.Settings.General.AccountSetup {
		r.Get(FormSetupEndpoint, SetupFormHandler)
		r.Post(FormSetupEndpoint, SetupHandler)
	}
}

// LoginFormHandler renders route GET "/login"
func LoginFormHandler(w http.ResponseWriter, r *http.Request) {
	data := models.QPFormLoginData{PageTitle: "Login"}

	templates := template.Must(template.ParseFiles(GetViewPath("layouts/main.tmpl"), GetViewPath("login.tmpl")))
	templates.ExecuteTemplate(w, "main", data)
}

// LoginHandler renders route POST "/login"
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
