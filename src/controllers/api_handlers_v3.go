package controllers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	library "github.com/nocodeleaks/quepasa/library"
	metrics "github.com/nocodeleaks/quepasa/metrics"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

const APIVersion3 string = "v3"

var ControllerPrefixV3 string = fmt.Sprintf("/%s/bot/{token}", APIVersion3)

func RegisterAPIV3Controllers(r chi.Router) {
	r.Get(ControllerPrefixV3, InformationControllerV3)

	// SENDING MSG ----------------------------
	// ----------------------------------------

	// used to dispatch alert msgs via url, triggers on monitor systems like zabbix
	r.Get(ControllerPrefixV3+"/send", SendAny)

	r.Post(ControllerPrefixV3+"/send", SendAny)
	r.Post(ControllerPrefixV3+"/send/{chatid}", SendAny)
	r.Post(ControllerPrefixV3+"/sendtext", SendText)
	r.Post(ControllerPrefixV3+"/sendtext/{chatid}", SendText)

	// SENDING MSG ATTACH ---------------------

	// deprecated, discard/remove on next version
	r.Post(ControllerPrefixV3+"/senddocument", SendDocumentAPIHandlerV2)

	r.Post(ControllerPrefixV3+"/sendurl", SendDocumentFromUrl)
	r.Post(ControllerPrefixV3+"/sendbinary/{chatid}/{filename}/{text}", SendDocumentFromBinary)
	r.Post(ControllerPrefixV3+"/sendbinary/{chatid}/{filename}", SendDocumentFromBinary)
	r.Post(ControllerPrefixV3+"/sendbinary/{chatid}", SendDocumentFromBinary)
	r.Post(ControllerPrefixV3+"/sendbinary", SendDocumentFromBinary)
	r.Post(ControllerPrefixV3+"/sendencoded", SendDocumentFromEncoded)

	// ----------------------------------------
	// SENDING MSG ----------------------------

	r.Get(ControllerPrefixV3+"/receive", ReceiveAPIHandler)
	r.Post(ControllerPrefixV3+"/attachment", AttachmentAPIHandlerV2)

	r.Get(ControllerPrefixV3+"/download/{messageid}", DownloadControllerV3)
	r.Get(ControllerPrefixV3+"/download", DownloadControllerV3)

	// PICTURE INFO | DATA --------------------
	// ----------------------------------------

	r.Post(ControllerPrefixV3+"/picinfo", PictureController)
	r.Get(ControllerPrefixV3+"/picinfo/{chatid}/{pictureid}", PictureController)
	r.Get(ControllerPrefixV3+"/picinfo/{chatid}", PictureController)
	r.Get(ControllerPrefixV3+"/picinfo", PictureController)

	r.Post(ControllerPrefixV3+"/picdata", PictureController)
	r.Get(ControllerPrefixV3+"/picdata/{chatid}/{pictureid}", PictureController)
	r.Get(ControllerPrefixV3+"/picdata/{chatid}", PictureController)
	r.Get(ControllerPrefixV3+"/picdata", PictureController)

	// ----------------------------------------
	// PICTURE INFO | DATA --------------------

	r.Post(ControllerPrefixV3+"/webhook", WebhookController)
	r.Get(ControllerPrefixV3+"/webhook", WebhookController)
	r.Delete(ControllerPrefixV3+"/webhook", WebhookController)

	// INVITE METHODS ************************
	// ----------------------------------------

	r.Get(ControllerPrefixV3+"/invite/{chatid}", InviteController)

	// ----------------------------------------
	// INVITE METHODS ************************
}

//region CONTROLLER - INFORMATION

// InformationController renders route GET "/{version}/info"
func InformationControllerV3(w http.ResponseWriter, r *http.Request) {
	// setting default reponse type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpInfoResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusNoContent)
		return
	}

	response.ParseSuccess(server)
	RespondSuccess(w, response)
}

//endregion

//endregion
//region CONTROLLER - DOWNLOAD MESSAGE ATTACHMENT

/*
<summary>

	Renders route GET "/{{version}}/bot/{{token}}/download/{messageid}"

	Any of then, at this order of priority
	Path parameters: {messageid}
	Url parameters: ?messageid={messageid} || ?id={messageid}
	Header parameters: X-QUEPASA-MESSAGEID = {messageid}

</summary>
*/
func DownloadControllerV3(w http.ResponseWriter, r *http.Request) {

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
