package controllers

import (
	"html/template"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"

	models "github.com/nocodeleaks/quepasa/models"
)

// Token of authentication / encryption
var TokenAuth = jwtauth.New("HS256", []byte(os.Getenv("SIGNING_SECRET")), nil)

// Prefix on forms endpoints to avoid conflict with api
const FormEndpointPrefix string = "/form"

var FormWebsocketEndpoint string = FormEndpointPrefix + "/verify/ws"
var FormAccountEndpoint string = FormEndpointPrefix + "/account"
var FormVerifyEndpoint string = FormEndpointPrefix + "/verify"
var FormDeleteEndpoint string = FormEndpointPrefix + "/delete"

func RegisterFormAuthenticatedControllers(r chi.Router) {
	r.Use(jwtauth.Verifier(TokenAuth))
	r.Use(HttpAuthenticatorHandler)

	r.HandleFunc(FormWebsocketEndpoint, VerifyHandler)
	r.Get(FormAccountEndpoint, FormAccountController)
	r.Get(FormVerifyEndpoint, VerifyFormHandler)

	r.Post(FormDeleteEndpoint, FormDeleteController)
	r.Post(FormEndpointPrefix+"/cycle", FormCycleController)
	r.Post(FormEndpointPrefix+"/debug", FormDebugController)
	r.Post(FormEndpointPrefix+"/toggle", FormToggleController)
	r.Post(FormEndpointPrefix+"/togglegroups", FormToggleGroupsController)
	r.Post(FormEndpointPrefix+"/togglebroadcast", FormToggleBroadcastController)

	r.Get(FormEndpointPrefix+"/server/{token}", FormSendController)
	r.Get(FormEndpointPrefix+"/server/{token}/send", FormSendController)
	r.Post(FormEndpointPrefix+"/server/{token}/send", FormSendController)
	r.Get(FormEndpointPrefix+"/server/{token}/receive", FormReceiveController)
}

// Authentication manager on forms
func HttpAuthenticatorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())

		if err != nil {
			http.Redirect(w, r, FormLoginEndpoint, http.StatusFound)
			return
		}

		if token == nil || !token.Valid {
			http.Redirect(w, r, FormLoginEndpoint, http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Rrenders route GET "/{prefix}/account"
func FormAccountController(w http.ResponseWriter, r *http.Request) {
	user, err := models.GetFormUser(r)
	if err != nil {
		RedirectToLogin(w, r)
	}

	data := models.QPFormAccountData{
		PageTitle: "Account",
		User:      *user,
	}

	data.Servers = models.GetServersForUser(user)
	data.Version = models.QPVersion
	templates := template.Must(template.ParseFiles("views/layouts/main.tmpl", "views/account.tmpl"))
	templates.ExecuteTemplate(w, "main", data)
}
