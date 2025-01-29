package controllers

import (
	"fmt"
	"net/http"
	"strings"

	models "github.com/nocodeleaks/quepasa/models"
)

//region CONTROLLER - INVITE

func InviteController(w http.ResponseWriter, r *http.Request) {

	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpInviteResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	chatId := models.GetChatId(r)
	if len(chatId) == 0 {
		err = fmt.Errorf("chat id missing")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	if !strings.HasSuffix(chatId, "@g.us") {
		err = fmt.Errorf("chatId must be a valid and formatted (@g.us) group id")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	switch os := r.Method; os {
	default:
		url, err := server.GetInvite(chatId)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		response.Url = url
		RespondSuccess(w, response)
		return
	}
}

//endregion
