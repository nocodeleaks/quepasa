package form

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"

	api "github.com/nocodeleaks/quepasa/api"
	environment "github.com/nocodeleaks/quepasa/environment"
	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	signalr "github.com/nocodeleaks/quepasa/signalr"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"
)

// GetFormEndpointPrefix returns the configured prefix for form endpoints
func GetFormEndpointPrefix() string {
	return "/" + environment.Settings.Form.Prefix
}

// Prefix on forms endpoints to avoid conflict with api
var FormEndpointPrefix string = GetFormEndpointPrefix()

var FormWebsocketEndpoint string = FormEndpointPrefix + "/verify/ws"
var FormAccountEndpoint string = FormEndpointPrefix + "/account"
var FormWebHooksEndpoint string = FormEndpointPrefix + "/webhooks"
var FormRabbitMQEndpoint string = FormEndpointPrefix + "/rabbitmq"
var FormVerifyEndpoint string = FormEndpointPrefix + "/verify"
var FormDeleteEndpoint string = FormEndpointPrefix + "/delete"

func RegisterFormAuthenticatedControllers(r chi.Router) {

	tokenAuth := GetTokenAuth()
	r.Use(jwtauth.Verifier(tokenAuth))

	r.Use(HttpAuthenticatorHandler)

	r.HandleFunc(FormWebsocketEndpoint, VerifyHandler)
	r.Get(FormAccountEndpoint, FormAccountController)
	r.Get(FormWebHooksEndpoint, FormWebHooksController)
	r.Get(FormRabbitMQEndpoint, FormRabbitMQController)
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

	masterkey := environment.Settings.API.MasterKey
	data.HasMasterKey = len(masterkey) > 0
	if data.HasMasterKey {
		data.HasSignalRActiveConnections = signalr.SignalRHub.HasActiveConnections(masterkey)
	}

	data.Servers = models.GetServersForUser(user)
	data.Version = models.QpVersion
	templates := template.Must(template.ParseFiles(GetViewPath("layouts/main.tmpl"), GetViewPath("account.tmpl")))
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
				data.Webhooks = server.GetWebhooks()
			}
		}
	} else {
		data.ErrorMessage = "missing token"
	}

	templates := template.Must(template.ParseFiles(
		GetViewPath("layouts/main.tmpl"),
		GetViewPath("webhooks.tmpl"),
	))

	templates.ExecuteTemplate(w, "main", data)
}

// Controller responsible for RabbitMQ management interface
func FormRabbitMQController(w http.ResponseWriter, r *http.Request) {
	// setting default response type as json
	w.Header().Set("Content-Type", "text/html")

	type FormRabbitMQControllerData struct {
		PageTitle    string                     `json:"pagetitle,omitempty"`
		ErrorMessage string                     `json:"errormessage,omitempty"`
		Server       *models.QpWhatsappServer   `json:"server,omitempty"`
		RabbitMQ     []*models.QpRabbitMQConfig `json:"rabbitmq,omitempty"`
	}

	data := FormRabbitMQControllerData{
		PageTitle: "RabbitMQ Configurations",
	}

	user, err := GetFormUser(r)
	if err == nil {
		token := r.URL.Query().Get("token")
		if len(token) > 0 {
			server, err := models.GetServerFromToken(token)
			if err != nil {
				data.ErrorMessage = "server token not found: " + err.Error()
			} else if server.User != user.Username {
				data.ErrorMessage = "server token not found or dont owned by you"
			} else {
				data.Server = server
				data.RabbitMQ = server.GetRabbitMQConfigsByQueue("")
			}
		}
	} else {
		data.ErrorMessage = "missing token"
	}

	templates := template.Must(template.ParseFiles(
		GetViewPath("layouts/main.tmpl"),
		GetViewPath("rabbitmq.tmpl"),
	))

	templates.ExecuteTemplate(w, "main", data)
}
