package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
)

const CurrentAPIVersion string = "v4"

func RegisterAPIControllers(r chi.Router) {

	// Basic health check route without authentication
	r.Get("/healthapi", BasicHealthController)

	aliases := []string{"/current", "", "/" + CurrentAPIVersion}
	for _, endpoint := range aliases {

		r.Get(endpoint+"/health", HealthController)
		r.Post(endpoint+"/account", AccountController)

		// CONTROL METHODS ************************
		// ----------------------------------------
		r.Get(endpoint+"/info", GetInformationController)
		r.Patch(endpoint+"/info", UpdateInformationController)
		r.Delete(endpoint+"/info", DeleteInformationController)

		r.Get(endpoint+"/scan", ScannerController)
		r.Get(endpoint+"/paircode", PairCodeController)

		r.Get(endpoint+"/command", CommandController)

		// ----------------------------------------
		// CONTROL METHODS ************************

		// SENDING MSG ----------------------------
		// ----------------------------------------

		r.Get(endpoint+"/message/{messageid}", GetMessageController)
		r.Get(endpoint+"/message", GetMessageController)

		r.Delete(endpoint+"/message/{messageid}", RevokeController)
		r.Delete(endpoint+"/message", RevokeController)

		// Mark message as read
		r.Post(endpoint+"/read", MarkReadController)

		// used to send alert msgs via url, triggers on monitor systems like zabbix
		r.Get(endpoint+"/send", SendAny)

		r.Post(endpoint+"/send", SendAny)
		r.Post(endpoint+"/send/{chatid}", SendAny)
		/*r.Post(endpoint+"/sendlinkpreview", SendWithLinkPreviewHandler)*/

		// obsolete, marked for remove (2024/10/22)
		r.Post(endpoint+"/sendtext", SendAny)
		r.Post(endpoint+"/sendtext/{chatid}", SendAny)

		// SENDING MSG ATTACH ---------------------

		// deprecated, discard/remove on next version
		r.Post(endpoint+"/senddocument", SendDocumentAPIHandlerV2)

		r.Post(endpoint+"/sendurl", SendAny)
		r.Post(endpoint+"/sendbinary/{chatid}/{filename}/{text}", SendDocumentFromBinary)
		r.Post(endpoint+"/sendbinary/{chatid}/{filename}", SendDocumentFromBinary)
		r.Post(endpoint+"/sendbinary/{chatid}", SendDocumentFromBinary)
		r.Post(endpoint+"/sendbinary", SendDocumentFromBinary)
		r.Post(endpoint+"/sendencoded", SendAny)

		// ----------------------------------------
		// SENDING MSG ----------------------------

		r.Get(endpoint+"/receive", ReceiveAPIHandler)
		r.Post(endpoint+"/attachment", AttachmentAPIHandlerV2)

		r.Get(endpoint+"/download/{messageid}", DownloadController)
		r.Get(endpoint+"/download", DownloadController)

		// PICTURE INFO | DATA --------------------
		// ----------------------------------------

		r.Post(endpoint+"/picinfo", PictureController)
		r.Get(endpoint+"/picinfo/{chatid}/{pictureid}", PictureController)
		r.Get(endpoint+"/picinfo/{chatid}", PictureController)
		r.Get(endpoint+"/picinfo", PictureController)

		r.Post(endpoint+"/picdata", PictureController)
		r.Get(endpoint+"/picdata/{chatid}/{pictureid}", PictureController)
		r.Get(endpoint+"/picdata/{chatid}", PictureController)
		r.Get(endpoint+"/picdata", PictureController)

		// ----------------------------------------
		// PICTURE INFO | DATA --------------------

		r.Post(endpoint+"/webhook", WebhookController)
		r.Get(endpoint+"/webhook", WebhookController)
		r.Delete(endpoint+"/webhook", WebhookController)

		// RABBITMQ DISPATCHING *******************
		// ----------------------------------------

		r.Post(endpoint+"/rabbitmq", RabbitMQController)
		r.Get(endpoint+"/rabbitmq", RabbitMQController)
		r.Delete(endpoint+"/rabbitmq", RabbitMQController)

		// ----------------------------------------
		// RABBITMQ DISPATCHING *******************

		// INVITE METHODS ************************
		// ----------------------------------------

		r.Get(endpoint+"/invite", InviteController)
		r.Get(endpoint+"/invite/{chatid}", InviteController)

		// ----------------------------------------
		// INVITE METHODS ************************

		r.Get(endpoint+"/contacts", ContactsController)
		r.Post(endpoint+"/isonwhatsapp", IsOnWhatsappController)

		// LID METHODS ****************************
		// ----------------------------------------

		r.Get(endpoint+"/useridentifier", GetUserIdentifierController)

		r.Get(endpoint+"/getphone", GetPhoneController)

		// ----------------------------------------
		// LID METHODS ****************************

		// USER INFO METHODS **********************
		// ----------------------------------------

		r.Post(endpoint+"/userinfo", UserInfoController)

		// ----------------------------------------
		// USER INFO METHODS **********************

		// IF YOU LOVE YOUR FREEDOM, DO NOT USE THAT
		// IT WAS DEVELOPED IN A MOMENT OF WEAKNESS
		// DONT BE THAT GUY !
		r.Post(endpoint+"/spam", Spam)

		// GROUPS CONTROLLER **********************
		// ----------------------------------------

		// Get all groups
		r.Get(endpoint+"/groups/getall", FetchAllGroupsController)

		// Get group info
		r.Get(endpoint+"/groups/get", GetGroupController)

		// Create a new group.
		r.Post(endpoint+"/groups/create", CreateGroupController)

		// Leave group
		r.Post(endpoint+"/groups/leave", LeaveGroupController)

		// Updates the group name.
		r.Put(endpoint+"/groups/name", SetGroupNameController)

		// Updates the group description.
		r.Put(endpoint+"/groups/description", SetGroupTopicController)

		// Updates the group picture.
		r.Put(endpoint+"/groups/photo", SetGroupPhotoController)

		// Updates the group participants.
		r.Put(endpoint+"/groups/participants", UpdateGroupParticipantsController)

		// Get group join requests
		r.Get(endpoint+"/groups/requests", GroupMembershipRequestsController)

		// Manage group join requests
		r.Post(endpoint+"/groups/requests", GroupMembershipRequestsController)

		// ----------------------------------------
		// GROUPS CONTROLLER **********************

		// Typing Controller ********************
		// ----------------------------------------
		r.Post(endpoint+"/chat/presence", ChatPresenceController)

		// ----------------------------------------
		// Typing Controller ********************

		// CHAT READ STATUS CONTROLLER **********
		// ----------------------------------------
		r.Post(endpoint+"/chat/markread", MarkChatAsReadController)
		r.Post(endpoint+"/chat/markunread", MarkChatAsUnreadController)

		// ----------------------------------------
		// CHAT READ STATUS CONTROLLER **********

		// CHAT ARCHIVE CONTROLLER **************
		// ----------------------------------------
		r.Post(endpoint+"/chat/archive", ArchiveChatController)

		// ----------------------------------------
		// CHAT ARCHIVE CONTROLLER **************

		// MESSAGE EDITING CONTROLLER ***********
		// ----------------------------------------
		r.Put(endpoint+"/edit", EditMessageController)

		// ----------------------------------------
		// MESSAGE EDITING CONTROLLER ***********

	}
}

// CommandController manages bot server commands
//
//	@Summary		Execute bot commands
//	@Description	Execute control commands for the bot server (start, stop, restart, status)
//	@Tags			Bot
//	@Accept			json
//	@Produce		json
//	@Param			action	query		string	true	"Command action"	Enums(start, stop, restart, status)
//	@Success		200		{object}	models.QpResponse
//	@Failure		400		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/command [get]
func CommandController(w http.ResponseWriter, r *http.Request) {
	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	action := library.GetRequestParameter(r, "action")
	switch action {
	case "start":
		err = server.Start()
		if err == nil {
			response.ParseSuccess("started")
		}
	case "stop":
		err = server.Stop("command")
		if err == nil {
			response.ParseSuccess("stopped")
		}
	case "restart":
		err = server.Restart()
		if err == nil {
			response.ParseSuccess("restarted")
		}
	case "status":
		status := server.GetStatus()
		response.ParseSuccess(status.String())
	case "groups":
		err := models.ToggleGroups(server)
		if err == nil {
			message := "groups toggled: " + server.Groups.String()
			response.ParseSuccess(message)
		}
	case "broadcasts":
		err := models.ToggleBroadcasts(server)
		if err == nil {
			message := "broadcasts toggled: " + server.Broadcasts.String()
			response.ParseSuccess(message)
		}
	case "readreceipts":
		err := models.ToggleReadReceipts(server)
		if err == nil {
			message := "readreceipts toggled: " + server.ReadReceipts.String()
			response.ParseSuccess(message)
		}
	case "calls":
		err := models.ToggleCalls(server)
		if err == nil {
			message := "calls toggled: " + server.Calls.String()
			response.ParseSuccess(message)
		}
	default:
		err = fmt.Errorf("invalid action: {%s}, try {start,stop,restart,status,groups}", action)
	}

	if err != nil {
		response.ParseError(err)
	}

	RespondInterface(w, response)
}
