package api

import (
	"fmt"
	"net/http"

	library "github.com/nocodeleaks/quepasa/library"
	metrics "github.com/nocodeleaks/quepasa/metrics"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

//region CONTROLLER - DOWNLOAD

/*
<summary>

	Renders route GET "/download/{messageid}"

	Any of then, at this order of priority
	Path parameters: {messageid}
	Url parameters: ?messageid={messageid} || ?id={messageid}
	Header parameters: X-QUEPASA-MESSAGEID = {messageid}

</summary>
*/
// DownloadController downloads media files from messages
// @Summary Download media
// @Description Downloads media files (images, videos, documents) from WhatsApp messages
// @Tags Download
// @Produce application/octet-stream
// @Param messageid path string false "Message ID (path parameter)"
// @Param id query string false "Message ID (query parameter)"
// @Param messageid query string false "Message ID (query parameter alternate)"
// @Param cache query string false "Use cached content"
// @Param X-QUEPASA-MESSAGEID header string false "Message ID (header parameter)"
// @Success 200 {file} binary "Media file"
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /download/{messageid} [get]
func DownloadController(w http.ResponseWriter, r *http.Request) {

	response := &models.QpResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Checking for ready state
	status := server.GetStatus()
	if status != whatsapp.Ready {
		err = &ApiServerNotReadyException{Wid: server.GetWId(), Status: status}
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
		return
	}

	// Default parameters
	messageid := GetMessageId(r)

	if len(messageid) == 0 {
		metrics.MessageSendErrors.Inc()
		err := fmt.Errorf("empty message id")
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Default parameters
	cache := GetCache(r)

	att, err := server.Download(messageid, cache)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	filename := att.FileName

	// If filename not setted
	if len(filename) == 0 {
		exten, ok := library.TryGetExtensionFromMimeType(att.Mimetype)
		if ok {
			// Generate from mime type and message id
			filename = messageid + exten
		}
	}

	var disposition string
	if len(filename) > 0 {
		disposition = fmt.Sprintf("attachment; filename=%q", filename)
	} else {
		disposition = "attachment"
	}

	// setting header filename
	w.Header().Set("Content-Disposition", disposition)

	// setting custom header content type
	if len(att.Mimetype) > 0 {
		w.Header().Set("Content-Type", att.Mimetype)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(*att.GetContent())
}

//endregion
