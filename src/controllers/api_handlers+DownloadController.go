package controllers

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
		err = &ApiServerNotReadyException{Wid: server.GetWid(), Status: status}
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

	var filename string

	// If filename not setted
	if len(att.FileName) == 0 {
		exten, ok := library.TryGetExtensionFromMimeType(att.Mimetype)
		if ok {

			// Generate from mime type and message id
			filename = fmt.Sprint("; filename=", messageid, exten)
		}
	} else {
		filename = fmt.Sprint("; filename=", att.FileName)
	}

	// setting header filename
	w.Header().Set("Content-Disposition", fmt.Sprint("attachment", filename))

	w.WriteHeader(http.StatusOK)
	w.Write(*att.GetContent())
}

//endregion
