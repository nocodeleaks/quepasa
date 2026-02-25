package api

import (
	"net/http"
	"time"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// SPAServerEnableController enables (starts) a server
//
//	@Summary		Enable a server (SPA)
//	@Description	Starts the WhatsApp server instance identified by {token}. Sends a system message to notify connected listeners.
//	@Tags		SPA
//	@Accept		json
//	@Produce	json
//	@Param		token	path	string	true	"Server token"
//	@Success	200	{object}	models.QpResponse
//	@Failure	400	{object}	models.QpResponse
//	@Failure	401	{object}	models.QpResponse
//	@Failure	500	{object}	models.QpResponse
//	@Security	Bearer
//	@Router		/server/{token}/enable [post]
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

	// notify
	if server.Handler != nil {
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

// SPAServerDisableController disables (stops) a server
//
//	@Summary		Disable a server (SPA)
//	@Description	Stops the WhatsApp server instance identified by {token}. Sends feedback to the caller on success.
//	@Tags		SPA
//	@Accept		json
//	@Produce	json
//	@Param		token	path	string	true	"Server token"
//	@Success	200	{object}	models.QpResponse
//	@Failure	400	{object}	models.QpResponse
//	@Failure	401	{object}	models.QpResponse
//	@Failure	500	{object}	models.QpResponse
//	@Security	Bearer
//	@Router		/server/{token}/disable [post]
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
