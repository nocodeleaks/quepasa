package controllers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	websocket "github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	models "github.com/nocodeleaks/quepasa/models"
)

func RedirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, FormLoginEndpoint, http.StatusFound)
}

// Google chrome bloqueou wss, portanto retornaremos sempre ws apatir de agora
func WebSocketProtocol() string {
	protocol := "ws"
	isSecure := models.ENV.UseSSLForWebSocket()
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
		RespondServerError(server, w, err)
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
	key := models.GetRequestParameter(r, "key")
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
			url := models.GetRequestParameter(r, "url")
			webhook := server.GetWebHook(url)
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
		} else {
			err = fmt.Errorf("invalid key or prefix: %s", key)
		}
	} else {
		err = fmt.Errorf("missing toggle key")
	}

	if err != nil {
		RespondServerError(server, w, err)
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
	user, err := models.GetFormUser(r)
	if err != nil {
		if err == models.ErrFormUnauthenticated {
			RedirectToLogin(w, r)
			return
		}

		RespondInterface(w, err)
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

	HSDString := models.GetRequestParameter(r, "historysyncdays")
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
	key := models.GetRequestParameter(r, "key")
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
				url := models.GetRequestParameter(r, "url")
				var affected uint
				affected, err = server.WebhookRemove(url)
				if affected > 0 {
					logentry.Infof("webhook delete requested by from, affected rows: %v", affected)
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
		RespondServerError(server, w, err)
		return
	}

	http.Redirect(w, r, destination, http.StatusFound)
}

//
// Helpers
//

// Facilitador que traz usuario e servidor para quem esta autenticado
func GetUserAndServer(w http.ResponseWriter, r *http.Request) (user *models.QpUser, server *models.QpWhatsappServer, err error) {
	user, err = models.GetFormUser(r)
	if err != nil {
		if err == models.ErrFormUnauthenticated {
			RedirectToLogin(w, r)
			return
		}

		RespondInterface(w, err)
		return
	}

	r.ParseForm()

	token := GetToken(r)
	server, err = models.WhatsappService.FindByToken(token)
	if err != nil {
		err = fmt.Errorf("get user and server error: %s", err.Error())
		return
	}

	return
}

func GetServerFromRequest(r *http.Request) (server *models.QpWhatsappServer, err error) {
	token := GetToken(r)
	return models.WhatsappService.FindByToken(token)
}

func GetDownloadPrefix(token string) (path string) {
	path = "/download?token={token}&cache=false&messageid={messageid}"
	path = strings.Replace(path, "{token}", token, -1)
	path = strings.Replace(path, "{messageid}", "", -1)
	return
}
