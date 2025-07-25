package form

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"

	api "github.com/nocodeleaks/quepasa/api"
	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"
)

// Prefix on forms endpoints to avoid conflict with api
const FormEndpointPrefix string = "/form"

var FormWebsocketEndpoint string = FormEndpointPrefix + "/verify/ws"
var FormAccountEndpoint string = FormEndpointPrefix + "/account"
var FormWebHooksEndpoint string = FormEndpointPrefix + "/webhooks"
var FormVerifyEndpoint string = FormEndpointPrefix + "/verify"
var FormDeleteEndpoint string = FormEndpointPrefix + "/delete"

func RegisterFormAuthenticatedControllers(r chi.Router) {

	tokenAuth := GetTokenAuth()
	r.Use(jwtauth.Verifier(tokenAuth))

	r.Use(HttpAuthenticatorHandler)

	r.HandleFunc(FormWebsocketEndpoint, VerifyHandler)
	r.Get(FormAccountEndpoint, FormAccountController)
	r.Get(FormWebHooksEndpoint, FormWebHooksController)
	r.Get(FormVerifyEndpoint, VerifyFormHandler)

	r.Post(FormDeleteEndpoint, FormDeleteController)
	r.Post(FormEndpointPrefix+"/debug", FormDebugController)
	r.Post(FormEndpointPrefix+"/toggle", FormToggleController)

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

// Renders route GET "/{prefix}/account"
func FormAccountController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUser(r)
	if err != nil {
		if err == models.ErrFormUnauthenticated {
			RedirectToLogin(w, r)
			return
		}

		api.RespondInterface(w, err)
		return
	}

	data := models.QPFormAccountData{
		PageTitle: "Account",
		User:      *user,
		Options:   whatsapp.Options,
		WMOptions: whatsmeow.WhatsmeowService.Options,
	}

	masterkey := models.ENV.MasterKey()
	data.HasMasterKey = len(masterkey) > 0
	if data.HasMasterKey {
		data.HasSignalRActiveConnections = models.SignalRHub.HasActiveConnections(masterkey)
	}

	data.Servers = models.GetServersForUser(user)
	data.Version = models.QpVersion
	templates := template.Must(template.ParseFiles("views/layouts/main.tmpl", "views/account.tmpl"))
	templates.ExecuteTemplate(w, "main", data)
}

// Renders route GET "/{prefix}/account"
func FormWebHooksController(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUser(r)
	if err != nil {
		if err == models.ErrFormUnauthenticated {
			RedirectToLogin(w, r)
			return
		}

		api.RespondInterface(w, err)
		return
	}

	data := models.QPFormWebHooksData{PageTitle: "WebHooks"}

	token := library.GetRequestParameter(r, "token")
	if len(token) > 0 {
		server, err := models.GetServerFromToken(token)
		if err != nil {
			data.ErrorMessage = "server not found"
		} else {
			if server.User != user.Username {
				data.ErrorMessage = "server token not found or dont owned by you"
			} else {
				data.Server = server
			}
		}
	} else {
		data.ErrorMessage = "missing token"
	}

	templates := template.Must(template.ParseFiles(
		"views/layouts/main.tmpl",
		"views/webhooks.tmpl",
	))

	templates.ExecuteTemplate(w, "main", data)
}
