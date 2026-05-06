package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apiModels "github.com/nocodeleaks/quepasa/api/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	qpwhatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/util/random"
	"go.mau.fi/whatsmeow/proto/waE2E"
	types "go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// LIDDirectSendRequest is an explicit transport contract for testing direct send to @lid targets.
type LIDDirectSendRequest struct {
	// ChatId must be a WhatsApp LID, for example: 121281638842371@lid
	ChatId string `json:"chatid"`

	// Text is the message body.
	Text string `json:"text"`

	// Optional id of the quoted message.
	InReply string `json:"inreply,omitempty"`

	// Optional tracking id.
	TrackId string `json:"trackid,omitempty"`
}

// SendLIDDirectController sends text directly to a @lid JID.
// This endpoint bypasses the API-layer LID->phone conversion path used by standard /messages send.
//
//	@Summary		Send text directly to @lid
//	@Description	Testing endpoint that sends a text message directly to a @lid destination without converting recipient to phone/@s.whatsapp.net in API layer.
//	@Tags			Send
//	@Accept			json
//	@Produce		json
//	@Param			request	body		api.LIDDirectSendRequest	true	"Direct @lid send request"
//	@Success		200		{object}	api.SendResponse
//	@Failure		400		{object}	api.SendResponse
//	@Failure		503		{object}	api.SendResponse
//	@Security		ApiKeyAuth
//	@Router			/messages/lid/direct [post]
func SendLIDDirectController(w http.ResponseWriter, r *http.Request) {
	response := &apiModels.SendResponse{}

	server, err := GetServer(r)
	if err != nil {
		MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	if server.GetStatus() != whatsapp.Ready {
		err = &ApiServerNotReadyException{Wid: server.GetWId(), Status: server.GetStatus()}
		MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
		return
	}

	request := &LIDDirectSendRequest{}
	if err = json.NewDecoder(r.Body).Decode(request); err != nil {
		MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("invalid json body: %s", err.Error()))
		RespondInterface(w, response)
		return
	}

	request.ChatId = strings.TrimSpace(request.ChatId)
	request.Text = strings.TrimSpace(request.Text)

	if request.ChatId == "" {
		MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("chatid is required"))
		RespondInterface(w, response)
		return
	}

	if !strings.HasSuffix(strings.ToLower(request.ChatId), whatsapp.WHATSAPP_SERVERDOMAIN_LID_SUFFIX) {
		MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("chatid must be a @lid destination"))
		RespondInterface(w, response)
		return
	}

	if request.Text == "" {
		MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("text is required"))
		RespondInterface(w, response)
		return
	}

	conn, err := server.GetValidConnection()
	if err != nil {
		MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	wmConn, ok := conn.(*qpwhatsmeow.WhatsmeowConnection)
	if !ok || wmConn == nil || wmConn.Client == nil {
		MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("active connection is not a valid whatsmeow client"))
		RespondInterface(w, response)
		return
	}

	lidJID, err := types.ParseJID(request.ChatId)
	if err != nil {
		MessageSendErrors.Inc()
		response.ParseError(fmt.Errorf("invalid @lid jid: %w", err))
		RespondInterface(w, response)
		return
	}

	waMsg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(request.Text),
		},
		MessageContextInfo: &waE2E.MessageContextInfo{
			MessageSecret: random.Bytes(32),
		},
	}

	sendResponse, err := wmConn.Client.SendMessage(context.Background(), lidJID, waMsg)
	if err != nil {
		log.Errorf("(%s) LIDDirectSend failed | chatid: %s | jid: %s | error: %v", server.GetWId(), request.ChatId, lidJID.String(), err)
		MessageSendErrors.Inc()
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	log.Infof("(%s) LIDDirectSend success | chatid: %s | jid: %s | msgid: %s", server.GetWId(), request.ChatId, lidJID.String(), sendResponse.ID)

	MessagesSent.Inc()

	result := &apiModels.SendResponseMessage{}
	result.Wid = server.GetWId()
	result.Id = sendResponse.ID
	result.ChatId = request.ChatId
	result.TrackId = request.TrackId

	response.ParseSuccess(result)
	RespondInterface(w, response)
}
