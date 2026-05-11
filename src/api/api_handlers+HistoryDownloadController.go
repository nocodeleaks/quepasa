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
	runtime "github.com/nocodeleaks/quepasa/runtime"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// AuthenticatedServerHistoryDownloadController downloads media referenced by a history-sync protocol message.
//
// WhatsApp history sync can arrive as protocol/debug events that point to media but
// do not yet carry a QuePasa attachment. This endpoint reconstructs enough metadata
// to reuse the normal connection download path, then patches the cached message so
// the web client can render and download the attachment like a regular message.
func AuthenticatedServerHistoryDownloadController(w http.ResponseWriter, r *http.Request) {
	response := &models.QpResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	if server.GetStatus() != whatsapp.Ready {
		err = &ApiServerNotReadyException{Wid: server.GetWId(), Status: server.GetStatus()}
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
		return
	}

	messageID := chi.URLParam(r, "messageid")
	if messageID == "" {
		response.ParseError(fmt.Errorf("missing messageid"))
		RespondInterfaceCode(w, response, http.StatusBadRequest)
		return
	}

	msg, err := server.Handler.GetById(messageID)
	if err != nil {
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusNotFound)
		return
	}

	if msg.Debug == nil || msg.Debug.Event != "ProtocolMessage" {
		response.ParseError(fmt.Errorf("message is not a protocol/history sync event"))
		RespondInterfaceCode(w, response, http.StatusBadRequest)
		return
	}

	switch debug := msg.Debug.Info.(type) {
	case *waE2E.ProtocolMessage:
		notif := debug.GetHistorySyncNotification()
		if notif == nil {
			response.ParseError(fmt.Errorf("no history sync notification inside protocol message"))
			RespondInterfaceCode(w, response, http.StatusBadRequest)
			return
		}

		directPath := notif.GetDirectPath()
		waMsg, err := buildHistorySyncDownloadMessage(directPath, notif)
		if err != nil {
			response.ParseError(err)
			RespondInterfaceCode(w, response, http.StatusBadRequest)
			return
		}

		// Re-wrap the protocol notification as a temporary WhatsApp message because
		// the connection download API expects an IWhatsappMessage-shaped payload.
		tmp := &whatsapp.WhatsappMessage{
			Id:          msg.Id + "-history-download",
			Timestamp:   time.Now().UTC(),
			Type:        whatsapp.UnhandledMessageType,
			Chat:        msg.Chat,
			FromHistory: true,
			Content:     waMsg,
		}

		conn, err := server.GetValidConnection()
		if err != nil {
			response.ParseError(fmt.Errorf("server connection unavailable: %v", err))
			RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
			return
		}

		downloaded, err := conn.DownloadData(tmp)
		if err != nil {
			response.ParseError(fmt.Errorf("failed to download history media: %v", err))
			RespondInterfaceCode(w, response, http.StatusInternalServerError)
			return
		}

		if len(downloaded) == 0 {
			response.ParseError(fmt.Errorf("download returned empty content"))
			RespondInterfaceCode(w, response, http.StatusInternalServerError)
			return
		}

		// Attach the downloaded bytes to the original cached message so future reads
		// and dispatcher notifications see the enriched version.
		attachment := &whatsapp.WhatsappAttachment{
			FileLength: uint64(len(downloaded)),
			FileName:   path.Base(directPath),
		}
		attachment.SetContent(&downloaded)

		if ext := path.Ext(attachment.FileName); ext != "" {
			attachment.Mimetype = mime.TypeByExtension(ext)
		}
		if attachment.Mimetype == "" {
			attachment.Mimetype = http.DetectContentType(downloaded)
		}
		if attachment.Mimetype == "" {
			attachment.Mimetype = "application/octet-stream"
		}

		if prefix, err := runtime.GetSessionDownloadPrefix(server.Token); err == nil {
			attachment.Url = prefix + "/" + msg.Id
		}

		if notif.GetFileSHA256() != nil {
			attachment.Checksum = fmt.Sprintf("%x", notif.GetFileSHA256())
		}

		msg.Attachment = attachment
		if isHistorySyncImagePath(directPath) {
			msg.Type = whatsapp.ImageMessageType
		} else {
			msg.Type = whatsapp.DocumentMessageType
		}

		if server.Handler != nil {
			// Trigger a synthetic update event so connected clients can refresh the
			// message card without waiting for another sync cycle.
			server.Handler.Trigger(msg)
		}

		RespondSuccess(w, map[string]interface{}{"result": "success", "id": msg.Id})
		return

	default:
		var payload map[string]interface{}
		body, _ := json.Marshal(msg.Debug.Info)
		_ = json.Unmarshal(body, &payload)
		if _, ok := payload["historySyncNotification"]; ok {
			response.ParseError(fmt.Errorf("protocol message present but cannot be processed on server side"))
			RespondInterfaceCode(w, response, http.StatusBadRequest)
			return
		}

		response.ParseError(fmt.Errorf("unsupported debug.info type for history download"))
		RespondInterfaceCode(w, response, http.StatusBadRequest)
		return
	}
}

// buildHistorySyncDownloadMessage converts a history-sync media descriptor into the
// protobuf shape expected by the active WhatsApp connection download routine.
func buildHistorySyncDownloadMessage(directPath string, notif *waE2E.HistorySyncNotification) (*waE2E.Message, error) {
	if directPath == "" {
		return nil, fmt.Errorf("missing direct path in history sync notification")
	}

	if isHistorySyncImagePath(directPath) {
		return &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				DirectPath:    proto.String(directPath),
				MediaKey:      notif.MediaKey,
				FileEncSHA256: notif.GetFileEncSHA256(),
				FileSHA256:    notif.GetFileSHA256(),
				FileLength:    proto.Uint64(notif.GetFileLength()),
			},
		}, nil
	}

	return &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			DirectPath:    proto.String(directPath),
			MediaKey:      notif.MediaKey,
			FileEncSHA256: notif.GetFileEncSHA256(),
			FileSHA256:    notif.GetFileSHA256(),
			FileLength:    proto.Uint64(notif.GetFileLength()),
		},
	}, nil
}

// isHistorySyncImagePath uses the direct path as a lightweight heuristic to decide
// whether the downloaded blob should be exposed as image or generic document media.
func isHistorySyncImagePath(directPath string) bool {
	lowerPath := strings.ToLower(directPath)
	return strings.Contains(lowerPath, ".jpg") ||
		strings.Contains(lowerPath, ".jpeg") ||
		strings.Contains(lowerPath, ".png") ||
		strings.Contains(lowerPath, ".webp")
}
