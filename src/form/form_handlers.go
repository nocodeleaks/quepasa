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
)

var FormLoginEndpoint string = "/login"
var FormSetupEndpoint string = "/setup"
var FormLogoutEndpoint string = "/logout"
var FormDownloadEndpoint string = "/download"

func RegisterFormControllers(r chi.Router) {

	// The SPA replaces server rendered pages. Redirect page GETs to the SPA root
	r.Get(FormLoginEndpoint, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusFound)
	})
	// Keep POST /login for API login; LoginHandler will return JSON and set cookie
	r.Post(FormLoginEndpoint, LoginHandler)
	r.Get(FormLogoutEndpoint, LogoutHandler)
	// Public JSON for login page customization
	r.Get(FormEndpointPrefix+"/login/json", FormLoginJSONController)

	// disable /setup if environment is false
	if environment.Settings.General.AccountSetup {
		r.Get(FormSetupEndpoint, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/", http.StatusFound)
		})
		r.Post(FormSetupEndpoint, SetupHandler)
	}
}

// LoginFormHandler renders route GET "/login"
func LoginFormHandler(w http.ResponseWriter, r *http.Request) {
	data := models.QPFormLoginData{PageTitle: "Login - Quepasa", Version: models.QpVersion}

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

	http.SetCookie(w, cookie)

	// For SPA clients, return a JSON success so frontend can proceed without expecting a redirect
	resp := map[string]string{"result": "success"}
	api.RespondSuccess(w, resp)
}
