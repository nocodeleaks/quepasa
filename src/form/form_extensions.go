package form

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/jwtauth"
	websocket "github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	api "github.com/nocodeleaks/quepasa/api"
	environment "github.com/nocodeleaks/quepasa/environment"
	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
)

func RedirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, FormLoginEndpoint, http.StatusFound)
}

// Google chrome bloqueou wss, portanto retornaremos sempre ws apatir de agora
func WebSocketProtocol() string {
	protocol := "ws"
	isSecure := environment.Settings.API.UseSSLWebSocket
	if isSecure {
		protocol = "wss"
	}

	return protocol
}

// DebugHandler renders route POST "/bot/debug"
func FormDebugController(w http.ResponseWriter, r *http.Request) {
	_, server, err := GetUserAndServer(w, r)
	if err != nil {
		// retorno já tratado pela funcao
		return
	}

	_, err = server.ToggleDevel()
	if err != nil {
		api.RespondServerError(server, w, err)
		return
	}

	http.Redirect(w, r, FormAccountEndpoint, http.StatusFound)
}

//#region TOGGLE

// ToggleHandler renders route POST "/form/toggle"
func FormToggleController(w http.ResponseWriter, r *http.Request) {
	var err error

	_, server, err := GetUserAndServer(w, r)
	if err != nil {
		return
	}

	destination := FormAccountEndpoint
	key := library.GetRequestParameter(r, "key")
	if len(key) > 0 {

		if strings.HasPrefix(key, "server") {

			switch key {
			case "server":
				{
					err = server.Toggle()
					break
				}
			case "server-broadcasts":
				{
					err = models.ToggleBroadcasts(server)
					break
				}
			case "server-groups":
				{
					err = models.ToggleGroups(server)
					break
				}
			case "server-readreceipts":
				{
					err = models.ToggleReadReceipts(server)
					break
				}
			case "server-calls":
				{
					err = models.ToggleCalls(server)
					break
				}

			default:
				{
					err = fmt.Errorf("invalid server key: %s", key)
					break
				}
			}
		} else if strings.HasPrefix(key, "webhook") {
			destination = FormWebHooksEndpoint + "?token=" + server.Token
			url := library.GetRequestParameter(r, "url")
			// Get webhook via dispatching system
			dispatching := server.GetDispatching(url)
			var webhook *models.QpWhatsappServerDispatching = nil
			if dispatching != nil && dispatching.IsWebhook() {
				webhook = models.NewQpWhatsappServerDispatchingFromDispatching(dispatching, server)
			}
			if webhook != nil {
				switch key {
				case "webhook-forwardinternal":
					{
						_, err = webhook.ToggleForwardInternal()
						break
					}
				case "webhook-broadcasts":
					{
						err = models.ToggleBroadcasts(webhook)
						break
					}
				case "webhook-groups":
					{
						err = models.ToggleGroups(webhook)
						break
					}
				case "webhook-readreceipts":
					{
						err = models.ToggleReadReceipts(webhook)
						break
					}
				case "webhook-calls":
					{
						err = models.ToggleCalls(webhook)
						break
					}
				default:
					{
						err = fmt.Errorf("invalid webhook key: %s", key)
						break
					}
				}
			} else {
				err = fmt.Errorf("webhook not found for url: %s", url)
			}
		} else if strings.HasPrefix(key, "rabbitmq") {
			destination = FormRabbitMQEndpoint + "?token=" + server.Token
			connectionString := library.GetRequestParameter(r, "connection_string")
			// Get RabbitMQ configuration via dispatching system
			dispatching := server.GetDispatching(connectionString)
			var rabbitmq *models.QpWhatsappServerDispatching = nil
			if dispatching != nil && dispatching.IsRabbitMQ() {
				// Use the new method that preserves the original type
				rabbitmq = models.NewQpWhatsappServerDispatchingFromDispatching(dispatching, server)
			}
			if rabbitmq != nil {
				switch key {
				case "rabbitmq-forwardinternal":
					{
						_, err = rabbitmq.ToggleForwardInternal()
						break
					}
				case "rabbitmq-broadcasts":
					{
						err = models.ToggleBroadcasts(rabbitmq)
						break
					}
				case "rabbitmq-groups":
					{
						err = models.ToggleGroups(rabbitmq)
						break
					}
				case "rabbitmq-readreceipts":
					{
						err = models.ToggleReadReceipts(rabbitmq)
						break
					}
				case "rabbitmq-calls":
					{
						err = models.ToggleCalls(rabbitmq)
						break
					}
				default:
					{
						err = fmt.Errorf("invalid rabbitmq key: %s", key)
						break
					}
				}
			} else {
				err = fmt.Errorf("rabbitmq configuration not found for connection: %s", connectionString)
			}
		} else {
			err = fmt.Errorf("invalid key or prefix: %s", key)
		}
	} else {
		err = fmt.Errorf("missing toggle key")
	}

	if err != nil {
		api.RespondServerError(server, w, err)
		return
	}

	http.Redirect(w, r, destination, http.StatusFound)
}

//#endregion

//
// Verify
//

// VerifyFormHandler renders route GET "/bot/verify" ?mode={sd|md}
func VerifyFormHandler(w http.ResponseWriter, r *http.Request) {
	data := models.QPFormVerifyData{
		PageTitle:   "Verify To Add or Update",
		Protocol:    WebSocketProtocol(),
		Host:        r.Host,
		Destination: FormAccountEndpoint,
	}

	templates := template.Must(template.ParseFiles(
		"views/layouts/main.tmpl",
		"views/bot/verify.tmpl",
	))
	templates.ExecuteTemplate(w, "main", data)
}

// VerifyHandler renders route GET "/bot/verify/ws"
func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetFormUser(r)
	if err != nil {
		if err == models.ErrFormUnauthenticated {
			RedirectToLogin(w, r)
			return
		}

		api.RespondInterface(w, err)
		return
	}

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("(websocket): service error: %s", err.Error())
		return
	}

	HSDString := library.GetRequestParameter(r, "historysyncdays")
	historysyncdays, _ := strconv.ParseUint(HSDString, 10, 32)

	pairing := &models.QpWhatsappPairing{
		Username:        user.Username,
		HistorySyncDays: uint32(historysyncdays),
	}

	WebSocketStart(*pairing, conn)
}

//
// Delete
//

// DeleteHandler renders route POST "/form/delete"
func FormDeleteController(w http.ResponseWriter, r *http.Request) {
	var err error

	_, server, err := GetUserAndServer(w, r)
	if err != nil {
		return
	}

	var destination string
	key := library.GetRequestParameter(r, "key")
	if len(key) > 0 {
		logentry := server.GetLogger()

		switch key {
		case "server":
			{
				destination = FormAccountEndpoint
				logentry.Warnf("delete requested by form !")
				err = models.WhatsappService.Delete(server)
			}
		case "webhook":
			{
				destination = FormWebHooksEndpoint + "?token=" + server.Token
				url := library.GetRequestParameter(r, "url")
				var affected uint
				affected, err = server.DispatchingRemove(url)
				if affected > 0 {
					logentry.Infof("webhook delete requested by from, affected rows: %v", affected)
				}
			}
		case "rabbitmq":
			{
				destination = FormRabbitMQEndpoint + "?token=" + server.Token
				connectionString := library.GetRequestParameter(r, "connection_string")
				if connectionString == "" {
					connectionString = library.GetRequestParameter(r, "exchange_name") // fallback for compatibility
				}
				if connectionString == "" {
					err = fmt.Errorf("connection_string is required for rabbitmq delete")
					break
				}
				var affected uint
				affected, err = server.DispatchingRemove(connectionString)
				if affected > 0 {
					logentry.Infof("rabbitmq delete requested by form, connection_string=%s, affected rows: %v", connectionString, affected)
				}
			}
		default:
			{
				err = fmt.Errorf("invalid delete key: %s", key)
			}
		}
	} else {
		err = fmt.Errorf("missing toggle key")
	}

	if err != nil {
		api.RespondServerError(server, w, err)
		return
	}

	http.Redirect(w, r, destination, http.StatusFound)
}

//
// Helpers
//

// Facilitador que traz usuario e servidor para quem esta autenticado
func GetUserAndServer(w http.ResponseWriter, r *http.Request) (user *models.QpUser, server *models.QpWhatsappServer, err error) {
	user, err = GetFormUser(r)
	if err != nil {
		if err == models.ErrFormUnauthenticated {
			RedirectToLogin(w, r)
			return
		}

		api.RespondInterface(w, err)
		return
	}

	r.ParseForm()

	token := api.GetToken(r)
	server, err = models.WhatsappService.FindByToken(token)
	if err != nil {
		err = fmt.Errorf("get user and server error: %s", err.Error())
		return
	}

	return
}

func GetServerFromRequest(r *http.Request) (server *models.QpWhatsappServer, err error) {
	token := api.GetToken(r)
	return models.WhatsappService.FindByToken(token)
}

func GetDownloadPrefix(token string) (path string) {
	path = "/download?token={token}&cache=false&messageid={messageid}"
	path = strings.Replace(path, "{token}", token, -1)
	path = strings.Replace(path, "{messageid}", "", -1)
	return
}

// GetUser gets the user_id from the JWT and finds the
// corresponding user in the database
func GetFormUser(r *http.Request) (*models.QpUser, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return nil, err
	}

	user, ok := claims["user_id"].(string)
	if !ok {
		return nil, models.ErrFormUnauthenticated
	}

	return models.WhatsappService.DB.Users.Find(user)
}
