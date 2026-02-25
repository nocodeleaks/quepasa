package api

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// SPAServerHistoryDownloadController downloads media referenced by a history-sync ProtocolMessage
//
//	@Summary		Download history-sync media
//	@Description	Given a ProtocolMessage history-sync event message id, attempts to download and decrypt the referenced media using the active Whatsmeow connection, attaches it to the message and triggers an update so the UI and dispatchers are notified.
//	@Tags		SPA
//	@Accept		json
//	@Produce	json
//	@Param		token		path	string	true	"Server token"
//	@Param		messageid	path	string	true	"Message id (protocol message)"
//	@Success	200	{object}	models.QpResponse
//	@Failure	400	{object}	models.QpResponse
//	@Failure	401	{object}	models.QpResponse
//	@Failure	403	{object}	models.QpResponse
//	@Failure	503	{object}	models.QpResponse
//	@Security	Bearer
//	@Router		/server/{token}/messages/{messageid}/history/download [post]
func SPAServerHistoryDownloadController(w http.ResponseWriter, r *http.Request) {
	response := &models.QpResponse{}

	// get server
	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// check readiness
	if server.GetStatus() != whatsapp.Ready {
		err := &ApiServerNotReadyException{Wid: server.GetWId(), Status: server.GetStatus()}
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
		return
	}

	messageId := chi.URLParam(r, "messageid")
	if messageId == "" {
		response.ParseError(fmt.Errorf("missing messageid"))
		RespondInterfaceCode(w, response, http.StatusBadRequest)
		return
	}

	msg, gerr := server.Handler.GetById(messageId)
	if gerr != nil {
		response.ParseError(gerr)
		RespondInterfaceCode(w, response, http.StatusNotFound)
		return
	}

	if msg.Debug == nil || msg.Debug.Event != "ProtocolMessage" {
		response.ParseError(fmt.Errorf("message is not a protocol/history sync event"))
		RespondInterfaceCode(w, response, http.StatusBadRequest)
		return
	}

	// try type assertion
	switch debug := msg.Debug.Info.(type) {
	case *waE2E.ProtocolMessage:
		notif := debug.GetHistorySyncNotification()
		if notif == nil {
			response.ParseError(fmt.Errorf("no history sync notification inside protocol message"))
			RespondInterfaceCode(w, response, http.StatusBadRequest)
			return
		}

		// Build a temporary waE2E.Message to allow download via existing connection.Download
		// We will pick document/image based on filename ext
		directPath := notif.GetDirectPath()
		var waMsg *waE2E.Message

		// naive extension check
		if strings.Contains(strings.ToLower(directPath), ".jpg") || strings.Contains(strings.ToLower(directPath), ".jpeg") || strings.Contains(strings.ToLower(directPath), ".png") || strings.Contains(strings.ToLower(directPath), ".webp") {
			internal := &waE2E.ImageMessage{
				DirectPath:    proto.String(directPath),
				MediaKey:      notif.MediaKey,
				FileEncSHA256: notif.GetFileEncSHA256(),
				FileSHA256:    notif.GetFileSHA256(),
				FileLength:    proto.Uint64(notif.GetFileLength()),
			}
			waMsg = &waE2E.Message{ImageMessage: internal}
		} else {
			internal := &waE2E.DocumentMessage{
				DirectPath:    proto.String(directPath),
				MediaKey:      notif.MediaKey,
				FileEncSHA256: notif.GetFileEncSHA256(),
				FileSHA256:    notif.GetFileSHA256(),
				FileLength:    proto.Uint64(notif.GetFileLength()),
			}
			waMsg = &waE2E.Message{DocumentMessage: internal}
		}

		// create temp wrapper whatsapp message
		tmp := &whatsapp.WhatsappMessage{
			Id:        msg.Id + "-history-download",
			Timestamp: time.Now().UTC(),
			Type:      whatsapp.UnhandledMessageType,
			Chat:      msg.Chat,
			FromHistory: true,
			Content:   waMsg,
		}

		// attempt download using active connection
		conn, cerr := server.GetValidConnection()
		if cerr != nil {
			response.ParseError(fmt.Errorf("server connection unavailable: %v", cerr))
			RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
			return
		}

		logentry := server.GetLogger().WithField("messageid", msg.Id)
		logentry.Infof("initiating history download for directPath=%s", directPath)

		downloaded, derr := conn.DownloadData(tmp)
		if derr != nil {
			logentry.Errorf("failed to download history media: %v", derr)
			response.ParseError(fmt.Errorf("failed to download history media: %v", derr))
			RespondInterfaceCode(w, response, http.StatusInternalServerError)
			return
		}
		logentry.Infof("downloaded %d bytes for message %s", len(downloaded), msg.Id)

		// attach to existing message
		attachment := &whatsapp.WhatsappAttachment{}
		// validate content
		if len(downloaded) == 0 {
			response.ParseError(fmt.Errorf("download returned empty content"))
			RespondInterfaceCode(w, response, http.StatusInternalServerError)
			return
		}
		attachment.SetContent(&downloaded)
		attachment.FileLength = uint64(len(downloaded))

		// derive filename and mimetype
		attachment.FileName = path.Base(directPath)
		if ext := path.Ext(attachment.FileName); len(ext) > 0 {
			attachment.Mimetype = mime.TypeByExtension(ext)
		}
		// sniff content-type as fallback
		if attachment.Mimetype == "" {
			attachment.Mimetype = http.DetectContentType(downloaded)
		}
		if attachment.Mimetype == "" {
			attachment.Mimetype = "application/octet-stream"
		}

		// set a public download url pointing to the standard /download endpoint
		if prefix, err := models.GetDownloadPrefixFromToken(server.Token); err == nil {
			attachment.Url = prefix + "/" + msg.Id
		}

		msg.Attachment = attachment

		// set checksum from notification sha256 when available
		if notif.GetFileSHA256() != nil {
			attachment.Checksum = fmt.Sprintf("%x", notif.GetFileSHA256())
		}

		// set message type based on extension (image vs generic document)
		if strings.Contains(strings.ToLower(directPath), ".jpg") || strings.Contains(strings.ToLower(directPath), ".jpeg") || strings.Contains(strings.ToLower(directPath), ".png") || strings.Contains(strings.ToLower(directPath), ".webp") {
			msg.Type = whatsapp.ImageMessageType
		} else {
			msg.Type = whatsapp.DocumentMessageType // generic
		}

		// Trigger update so UI and external dispatchers are notified
		if server.Handler != nil {
			server.Handler.Trigger(msg)
		}

		RespondSuccess(w, map[string]interface{}{"result": "success", "id": msg.Id})
		return

	default:
		// unknown debug content
		// try to unmarshal if it's a json map
		var m map[string]interface{}
		b, _ := json.Marshal(msg.Debug.Info)
		json.Unmarshal(b, &m)
		if _, ok := m["historySyncNotification"]; ok {
			// can't download without a connection -> return informative error
			response.ParseError(fmt.Errorf("protocol message present but cannot be processed on server side"))
			RespondInterfaceCode(w, response, http.StatusBadRequest)
			return
		}

		response.ParseError(fmt.Errorf("unsupported debug.info type for history download"))
		RespondInterfaceCode(w, response, http.StatusBadRequest)
		return
	}
}
