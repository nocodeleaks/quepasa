package legacy

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Config struct {
	CurrentAPIVersion string
	APIVersion3       string
}

type Handlers map[string]http.HandlerFunc

func RegisterAPIControllers(r chi.Router, config Config, handlers Handlers) {
	r.Get("/healthapi", mustHandler(handlers, "BasicHealthController"))

	aliases := []string{"/current", "", "/" + config.CurrentAPIVersion}
	for _, endpoint := range aliases {
		r.Get(endpoint+"/health", mustHandler(handlers, "HealthController"))
		r.Head(endpoint+"/health", mustHandler(handlers, "HealthController"))
		r.Get(endpoint+"/environment", mustHandler(handlers, "EnvironmentController"))
		r.Get(endpoint+"/login/config", mustHandler(handlers, "LoginConfigController"))
		r.Post(endpoint+"/account", mustHandler(handlers, "AccountController"))

		r.Post(endpoint+"/info", mustHandler(handlers, "CreateInformationController"))
		r.Get(endpoint+"/info", mustHandler(handlers, "GetInformationController"))
		r.Patch(endpoint+"/info", mustHandler(handlers, "UpdateInformationController"))
		r.Delete(endpoint+"/info", mustHandler(handlers, "DeleteInformationController"))

		r.Get(endpoint+"/scan", mustHandler(handlers, "ScannerController"))
		r.Get(endpoint+"/paircode", mustHandler(handlers, "PairCodeController"))
		r.Get(endpoint+"/command", mustHandler(handlers, "CommandController"))

		r.Get(endpoint+"/message/{messageid}", mustHandler(handlers, "GetMessageController"))
		r.Get(endpoint+"/message", mustHandler(handlers, "GetMessageController"))
		r.Delete(endpoint+"/message/{messageid}", mustHandler(handlers, "RevokeController"))
		r.Delete(endpoint+"/message", mustHandler(handlers, "RevokeController"))
		r.Post(endpoint+"/read", mustHandler(handlers, "MarkReadController"))

		r.Get(endpoint+"/send", mustHandler(handlers, "SendAny"))
		r.Post(endpoint+"/send", mustHandler(handlers, "SendAny"))
		r.Post(endpoint+"/send/{chatid}", mustHandler(handlers, "SendAny"))
		r.Post(endpoint+"/sendtext", mustHandler(handlers, "SendAny"))
		r.Post(endpoint+"/sendtext/{chatid}", mustHandler(handlers, "SendAny"))
		r.Post(endpoint+"/senddocument", mustHandler(handlers, "SendDocument"))
		r.Post(endpoint+"/senddocument/{chatid}", mustHandler(handlers, "SendDocument"))
		r.Post(endpoint+"/sendurl", mustHandler(handlers, "SendAny"))
		r.Post(endpoint+"/sendbinary/{chatid}/{filename}/{text}", mustHandler(handlers, "SendDocumentFromBinary"))
		r.Post(endpoint+"/sendbinary/{chatid}/{filename}", mustHandler(handlers, "SendDocumentFromBinary"))
		r.Post(endpoint+"/sendbinary/{chatid}", mustHandler(handlers, "SendDocumentFromBinary"))
		r.Post(endpoint+"/sendbinary", mustHandler(handlers, "SendDocumentFromBinary"))
		r.Post(endpoint+"/sendencoded", mustHandler(handlers, "SendAny"))

		r.Get(endpoint+"/receive", mustHandler(handlers, "ReceiveAPIHandler"))
		r.Post(endpoint+"/redispatch/{messageid}", mustHandler(handlers, "RedispatchAPIHandler"))
		r.Get(endpoint+"/download/{messageid}", mustHandler(handlers, "DownloadController"))
		r.Get(endpoint+"/download", mustHandler(handlers, "DownloadController"))

		r.Post(endpoint+"/picinfo", mustHandler(handlers, "PictureController"))
		r.Get(endpoint+"/picinfo/{chatid}/{pictureid}", mustHandler(handlers, "PictureController"))
		r.Get(endpoint+"/picinfo/{chatid}", mustHandler(handlers, "PictureController"))
		r.Get(endpoint+"/picinfo", mustHandler(handlers, "PictureController"))
		r.Post(endpoint+"/picdata", mustHandler(handlers, "PictureController"))
		r.Get(endpoint+"/picdata/{chatid}/{pictureid}", mustHandler(handlers, "PictureController"))
		r.Get(endpoint+"/picdata/{chatid}", mustHandler(handlers, "PictureController"))
		r.Get(endpoint+"/picdata", mustHandler(handlers, "PictureController"))

		r.Post(endpoint+"/webhook", mustHandler(handlers, "WebhookController"))
		r.Get(endpoint+"/webhook", mustHandler(handlers, "WebhookController"))
		r.Delete(endpoint+"/webhook", mustHandler(handlers, "WebhookController"))
		r.Post(endpoint+"/rabbitmq", mustHandler(handlers, "RabbitMQController"))
		r.Get(endpoint+"/rabbitmq", mustHandler(handlers, "RabbitMQController"))
		r.Delete(endpoint+"/rabbitmq", mustHandler(handlers, "RabbitMQController"))

		r.Get(endpoint+"/invite", mustHandler(handlers, "InviteController"))
		r.Get(endpoint+"/invite/{chatid}", mustHandler(handlers, "InviteController"))

		r.Get(endpoint+"/contacts", mustHandler(handlers, "ContactsController"))
		r.Post(endpoint+"/contact/search", mustHandler(handlers, "ContactSearchController"))
		r.Post(endpoint+"/isonwhatsapp", mustHandler(handlers, "IsOnWhatsappController"))
		r.Get(endpoint+"/useridentifier", mustHandler(handlers, "GetUserIdentifierController"))
		r.Get(endpoint+"/getphone", mustHandler(handlers, "GetPhoneController"))
		r.Post(endpoint+"/userinfo", mustHandler(handlers, "UserInfoController"))
		r.Post(endpoint+"/spam", mustHandler(handlers, "Spam"))

		r.Get(endpoint+"/groups/getall", mustHandler(handlers, "FetchAllGroupsController"))
		r.Get(endpoint+"/groups/get", mustHandler(handlers, "GetGroupController"))
		r.Post(endpoint+"/groups/create", mustHandler(handlers, "CreateGroupController"))
		r.Post(endpoint+"/groups/leave", mustHandler(handlers, "LeaveGroupController"))
		r.Put(endpoint+"/groups/name", mustHandler(handlers, "SetGroupNameController"))
		r.Put(endpoint+"/groups/description", mustHandler(handlers, "SetGroupTopicController"))
		r.Put(endpoint+"/groups/photo", mustHandler(handlers, "SetGroupPhotoController"))
		r.Put(endpoint+"/groups/participants", mustHandler(handlers, "UpdateGroupParticipantsController"))
		r.Get(endpoint+"/groups/requests", mustHandler(handlers, "GroupMembershipRequestsController"))
		r.Post(endpoint+"/groups/requests", mustHandler(handlers, "GroupMembershipRequestsController"))

		r.Post(endpoint+"/chat/presence", mustHandler(handlers, "ChatPresenceController"))
		r.Get(endpoint+"/labels", mustHandler(handlers, "ConversationLabelController"))
		r.Post(endpoint+"/labels", mustHandler(handlers, "ConversationLabelController"))
		r.Put(endpoint+"/labels", mustHandler(handlers, "ConversationLabelController"))
		r.Delete(endpoint+"/labels", mustHandler(handlers, "ConversationLabelController"))
		r.Get(endpoint+"/chat/labels", mustHandler(handlers, "ConversationChatLabelController"))
		r.Post(endpoint+"/chat/labels", mustHandler(handlers, "ConversationChatLabelController"))
		r.Delete(endpoint+"/chat/labels", mustHandler(handlers, "ConversationChatLabelController"))
		r.Post(endpoint+"/chat/markread", mustHandler(handlers, "MarkChatAsReadController"))
		r.Post(endpoint+"/chat/markunread", mustHandler(handlers, "MarkChatAsUnreadController"))
		r.Post(endpoint+"/chat/archive", mustHandler(handlers, "ArchiveChatController"))
		r.Put(endpoint+"/edit", mustHandler(handlers, "EditMessageController"))

		r.Get(endpoint+"/restore", mustHandler(handlers, "RestoreDiagnoseController"))
		r.Post(endpoint+"/restore/auto", mustHandler(handlers, "RestoreAutoController"))
		r.Post(endpoint+"/restore/manual", mustHandler(handlers, "RestoreManualController"))
	}
}

func RegisterAPIV3Controllers(r chi.Router, config Config, handlers Handlers) {
	controllerPrefixV3 := "/" + config.APIVersion3 + "/bot/{token}"

	r.Get(controllerPrefixV3, mustHandler(handlers, "InformationControllerV3"))
	r.Get(controllerPrefixV3+"/send", mustHandler(handlers, "SendAny"))
	r.Post(controllerPrefixV3+"/send", mustHandler(handlers, "SendAny"))
	r.Post(controllerPrefixV3+"/send/{chatid}", mustHandler(handlers, "SendAny"))
	r.Post(controllerPrefixV3+"/sendtext", mustHandler(handlers, "SendAny"))
	r.Post(controllerPrefixV3+"/sendtext/{chatid}", mustHandler(handlers, "SendAny"))
	r.Post(controllerPrefixV3+"/senddocument", mustHandler(handlers, "SendDocument"))
	r.Post(controllerPrefixV3+"/sendurl", mustHandler(handlers, "SendAny"))
	r.Post(controllerPrefixV3+"/sendbinary/{chatid}/{filename}/{text}", mustHandler(handlers, "SendDocumentFromBinary"))
	r.Post(controllerPrefixV3+"/sendbinary/{chatid}/{filename}", mustHandler(handlers, "SendDocumentFromBinary"))
	r.Post(controllerPrefixV3+"/sendbinary/{chatid}", mustHandler(handlers, "SendDocumentFromBinary"))
	r.Post(controllerPrefixV3+"/sendbinary", mustHandler(handlers, "SendDocumentFromBinary"))
	r.Post(controllerPrefixV3+"/sendencoded", mustHandler(handlers, "SendAny"))
	r.Get(controllerPrefixV3+"/receive", mustHandler(handlers, "ReceiveAPIHandler"))
	r.Get(controllerPrefixV3+"/download/{messageid}", mustHandler(handlers, "DownloadController"))
	r.Get(controllerPrefixV3+"/download", mustHandler(handlers, "DownloadController"))
	r.Post(controllerPrefixV3+"/picinfo", mustHandler(handlers, "PictureController"))
	r.Get(controllerPrefixV3+"/picinfo/{chatid}/{pictureid}", mustHandler(handlers, "PictureController"))
	r.Get(controllerPrefixV3+"/picinfo/{chatid}", mustHandler(handlers, "PictureController"))
	r.Get(controllerPrefixV3+"/picinfo", mustHandler(handlers, "PictureController"))
	r.Post(controllerPrefixV3+"/picdata", mustHandler(handlers, "PictureController"))
	r.Get(controllerPrefixV3+"/picdata/{chatid}/{pictureid}", mustHandler(handlers, "PictureController"))
	r.Get(controllerPrefixV3+"/picdata/{chatid}", mustHandler(handlers, "PictureController"))
	r.Get(controllerPrefixV3+"/picdata", mustHandler(handlers, "PictureController"))
	r.Post(controllerPrefixV3+"/webhook", mustHandler(handlers, "WebhookController"))
	r.Get(controllerPrefixV3+"/webhook", mustHandler(handlers, "WebhookController"))
	r.Delete(controllerPrefixV3+"/webhook", mustHandler(handlers, "WebhookController"))
	r.Get(controllerPrefixV3+"/labels", mustHandler(handlers, "ConversationLabelController"))
	r.Post(controllerPrefixV3+"/labels", mustHandler(handlers, "ConversationLabelController"))
	r.Put(controllerPrefixV3+"/labels", mustHandler(handlers, "ConversationLabelController"))
	r.Delete(controllerPrefixV3+"/labels", mustHandler(handlers, "ConversationLabelController"))
	r.Get(controllerPrefixV3+"/chat/labels", mustHandler(handlers, "ConversationChatLabelController"))
	r.Post(controllerPrefixV3+"/chat/labels", mustHandler(handlers, "ConversationChatLabelController"))
	r.Delete(controllerPrefixV3+"/chat/labels", mustHandler(handlers, "ConversationChatLabelController"))
	r.Get(controllerPrefixV3+"/invite/{chatid}", mustHandler(handlers, "InviteController"))
}

func mustHandler(handlers Handlers, name string) http.HandlerFunc {
	handler, ok := handlers[name]
	if !ok || handler == nil {
		panic("legacy route handler not configured: " + name)
	}
	return handler
}
