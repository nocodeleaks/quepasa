package api

import (
	"errors"
	"net/http"

	apiModels "github.com/nocodeleaks/quepasa/api/models"
	"github.com/nocodeleaks/quepasa/runtime"
)

var errSpamMasterKeyRequired = errors.New("master key required for spam endpoint")

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
//	@Success		200		{object}	api.SendResponse
//	@Failure		423		{object}	api.SendResponse	"No server available"
//	@Security		ApiKeyAuth
//	@Router			/spam [post]
func Spam(w http.ResponseWriter, r *http.Request) {
	if !IsMatchForMaster(r) {
		MessageSendErrors.Inc()

		response := &apiModels.SendResponse{}
		response.ParseError(errSpamMasterKeyRequired)
		RespondInterfaceCode(w, response, http.StatusLocked)
		return
	}

	server, err := runtime.GetSpamSession()
	if err != nil {
		MessageSendErrors.Inc()

		response := &apiModels.SendResponse{}
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusLocked)
		return
	}

	SendAnyWithServer(w, r, server)
}
