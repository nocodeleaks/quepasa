package api

import (
	"fmt"
	"net/http"
	"strings"

	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
)

//region CONTROLLER - INVITE

// InviteController generates invite links for WhatsApp groups
// @Summary Generate group invite link
// @Description Generates an invite link for a specific WhatsApp group
// @Tags Groups
// @Accept json
// @Produce json
// @Param chatid path string false "Chat ID (path parameter)"
// @Param chatid query string false "Chat ID (query parameter)"
// @Success 200 {object} models.QpInviteResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /invite/{chatid} [get]
// @Router /invite [get]
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

	chatId := library.GetChatId(r)
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
		url, err := server.GetGroupManager().GetInvite(chatId)
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
