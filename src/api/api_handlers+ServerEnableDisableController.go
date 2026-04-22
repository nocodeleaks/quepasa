package api

import (
	"net/http"
	"time"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// SPAServerEnableController starts a server through the SPA HTTP surface.
//
// The actual lifecycle logic stays in models.QpWhatsappServer.Start(). This handler
// only resolves the route token, translates errors into HTTP responses, and emits
// a lightweight system message so active listeners can refresh UI state quickly.
func SPAServerEnableController(w http.ResponseWriter, r *http.Request) {
	response := &models.QpResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	err = server.Start()
	if err != nil {
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusInternalServerError)
		return
	}

	if server.Handler != nil {
		// Mirror the state change as an in-memory system event so websocket-driven
		// clients do not have to poll immediately after a successful enable call.
		sysMsg := &whatsapp.WhatsappMessage{
			Id:        "server-enable-" + server.Token,
			Timestamp: time.Now().UTC(),
			Type:      whatsapp.SystemMessageType,
			FromMe:    false,
			Chat:      whatsapp.WASYSTEMCHAT,
			Text:      "Server enabled",
			Info:      map[string]interface{}{"event": "server_enabled"},
		}
		go server.Handler.Message(sysMsg, "server-enable-notify")
	}

	RespondSuccess(w, map[string]interface{}{"result": "started"})
}

// SPAServerDisableController stops a server through the SPA HTTP surface.
//
// Unlike enable, stop already emits internal lifecycle effects from the model layer,
// so this controller only forwards the request and standardizes the JSON response.
func SPAServerDisableController(w http.ResponseWriter, r *http.Request) {
	response := &models.QpResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	err = server.Stop("disabled via api")
	if err != nil {
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusInternalServerError)
		return
	}

	RespondSuccess(w, map[string]interface{}{"result": "stopped"})
}
