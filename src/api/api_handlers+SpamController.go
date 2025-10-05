package api

import (
	"net/http"

	metrics "github.com/nocodeleaks/quepasa/metrics"
	models "github.com/nocodeleaks/quepasa/models"
)

// -------------------------- PUBLIC METHODS
//region TYPES OF SPAMMING

// SendAPIHandler renders route "/v4/bot/{token}/spam"
// Returns 423 STATUS if no server available
//
//	@Summary		Send spam messages
//	@Description	Send messages using any available server (spam/broadcast functionality)
//	@Tags			Application
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{chatId=string,text=string}	true	"Spam message request"
//	@Success		200		{object}	models.QpSendResponse
//	@Failure		423		{object}	models.QpSendResponse	"No server available"
//	@Security		ApiKeyAuth
//	@Router			/spam [post]
func Spam(w http.ResponseWriter, r *http.Request) {
	server, err := GetServerFromMaster(r)
	if err != nil {
		metrics.MessageSendErrors.Inc()

		response := &models.QpSendResponse{}
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusLocked)
		return
	}

	SendAnyWithServer(w, r, server)
}
