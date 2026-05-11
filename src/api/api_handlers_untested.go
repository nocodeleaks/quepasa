package api

import (
	"net/http"

	apiModels "github.com/nocodeleaks/quepasa/api/models"
)

// SendDocumentFromBinary sends a document from binary data in the request body
func SendDocumentFromBinary(w http.ResponseWriter, r *http.Request) {
	response := &apiModels.SendResponse{}

	server, err := GetServer(r)
	if err != nil {
		MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Declare a new request struct.
	request := &apiModels.SendRequest{}

	// Getting ChatId parameter
	err = request.EnsureValidChatId(r)
	if err != nil {
		MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	err = request.GenerateBodyContent(r)
	if err != nil {
		MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	SendRequest(w, r, request, server)
}
