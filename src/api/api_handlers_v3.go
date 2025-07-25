package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	models "github.com/nocodeleaks/quepasa/models"
)

const APIVersion3 string = "v3"

var ControllerPrefixV3 string = fmt.Sprintf("/%s/bot/{token}", APIVersion3)

func RegisterAPIV3Controllers(r chi.Router) {

	r.Get(ControllerPrefixV3, InformationControllerV3)

	// SENDING MSG ----------------------------
	// ----------------------------------------

	// used to send alert msgs via url, triggers on monitor systems like zabbix
	r.Get(ControllerPrefixV3+"/send", SendAny)

	r.Post(ControllerPrefixV3+"/send", SendAny)
	r.Post(ControllerPrefixV3+"/send/{chatid}", SendAny)

	// obsolete, marked for remove (2024/10/22)
	r.Post(ControllerPrefixV3+"/sendtext", SendAny)
	r.Post(ControllerPrefixV3+"/sendtext/{chatid}", SendAny)

	// SENDING MSG ATTACH ---------------------

	// deprecated, discard/remove on next version
	r.Post(ControllerPrefixV3+"/senddocument", SendDocumentAPIHandlerV2)

	r.Post(ControllerPrefixV3+"/sendurl", SendAny)
	r.Post(ControllerPrefixV3+"/sendbinary/{chatid}/{filename}/{text}", SendDocumentFromBinary)
	r.Post(ControllerPrefixV3+"/sendbinary/{chatid}/{filename}", SendDocumentFromBinary)
	r.Post(ControllerPrefixV3+"/sendbinary/{chatid}", SendDocumentFromBinary)
	r.Post(ControllerPrefixV3+"/sendbinary", SendDocumentFromBinary)
	r.Post(ControllerPrefixV3+"/sendencoded", SendAny)

	// ----------------------------------------
	// SENDING MSG ----------------------------

	r.Get(ControllerPrefixV3+"/receive", ReceiveAPIHandler)
	r.Post(ControllerPrefixV3+"/attachment", AttachmentAPIHandlerV2)

	r.Get(ControllerPrefixV3+"/download/{messageid}", DownloadController)
	r.Get(ControllerPrefixV3+"/download", DownloadController)

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
	// setting default response type as json
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
