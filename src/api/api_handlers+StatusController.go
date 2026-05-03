package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// StatusPublishRequest defines the parameters for publishing a WhatsApp status (story).
type StatusPublishRequest struct {
	// Text is the caption for media status, or the full content for text-only status.
	Text string `json:"text"`

	// Attachment is optional. When provided, the status becomes a media story (image or video).
	// Leave nil to publish a text-only status.
	Attachment *whatsapp.WhatsappAttachment `json:"attachment,omitempty"`
}

// PublishStatusController publishes a WhatsApp status (story) visible to contacts.
//
//	@Summary		Publish WhatsApp status
//	@Description	Publishes a WhatsApp status (story). Send text for text-only status, or include an attachment for media status (image/video). The story expires after 24 hours on WhatsApp servers.
//	@Tags			Status
//	@Accept			json
//	@Produce		json
//	@Param			request	body		StatusPublishRequest	true	"Status publish request"
//	@Success		200		{object}	models.QpResponse
//	@Failure		400		{object}	models.QpResponse
//	@Failure		503		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/status/publish [post]
func PublishStatusController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	status := server.GetStatus()
	if status != whatsapp.Ready {
		err = &ApiServerNotReadyException{Wid: server.GetWId(), Status: status}
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
		return
	}

	var request StatusPublishRequest
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.ParseError(fmt.Errorf("invalid request body: %w", err))
		RespondInterface(w, response)
		return
	}

	if request.Text == "" && request.Attachment == nil {
		response.ParseError(fmt.Errorf("text or attachment is required"))
		RespondInterface(w, response)
		return
	}

	conn, err := server.GetValidConnection()
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	msgID, err := conn.PublishStatus(request.Text, request.Attachment)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.ParseSuccess(msgID)
	RespondInterface(w, response)
}
