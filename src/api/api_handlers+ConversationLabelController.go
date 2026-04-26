package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	apiModels "github.com/nocodeleaks/quepasa/api/models"
	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type conversationLabelRequest struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Active *bool  `json:"active,omitempty"`
}

type conversationChatLabelRequest struct {
	ChatID  string `json:"chatid"`
	LabelID int64  `json:"labelid"`
}

// ConversationLabelController manages the label catalog for the owner of the authenticated bot token.
//
//	@Summary		Manage conversation labels
//	@Description	Create, list, update and delete reusable conversation labels for the current bot owner
//	@Tags			Chat
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{id=integer,name=string,color=string,active=boolean}	false	"Label payload"
//	@Success		200		{object}	api.ConversationLabelsResponse
//	@Failure		400		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/labels [get]
//	@Router			/labels [post]
//	@Router			/labels [put]
//	@Router			/labels [delete]
func ConversationLabelController(w http.ResponseWriter, r *http.Request) {
	response := &apiModels.ConversationLabelsResponse{}

	serverRecord, err := getConversationLabelServerRecord(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	handleConversationLabelCatalog(w, r, strings.TrimSpace(serverRecord.GetUser()))
}

// ConversationChatLabelController manages label bindings for one conversation on the current bot token.
//
//	@Summary		Manage labels on one conversation
//	@Description	List, apply or remove labels from a specific conversation identified by chatid
//	@Tags			Chat
//	@Accept			json
//	@Produce		json
//	@Param			chatid	query		string	true	"Conversation chat id"
//	@Param			request	body		object{chatid=string,labelid=integer}	false	"Conversation label payload"
//	@Success		200		{object}	api.ConversationLabelsResponse
//	@Failure		400		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/chat/labels [get]
//	@Router			/chat/labels [post]
//	@Router			/chat/labels [delete]
func ConversationChatLabelController(w http.ResponseWriter, r *http.Request) {
	response := &apiModels.ConversationLabelsResponse{}

	serverRecord, err := getConversationLabelServerRecord(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	handleConversationChatLabels(w, r, strings.TrimSpace(serverRecord.GetUser()), strings.TrimSpace(serverRecord.Token))
}

// SPAConversationLabelController exposes the label catalog through SPA authentication.
func SPAConversationLabelController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	handleConversationLabelCatalog(w, r, strings.TrimSpace(user.Username))
}

// SPAServerConversationLabelController exposes conversation label bindings through SPA authentication.
func SPAServerConversationLabelController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetSPATokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	if _, err := GetSPAOwnedServerRecord(user, token); err != nil {
		respondSPAServerLookupError(w, err)
		return
	}

	handleConversationChatLabels(w, r, strings.TrimSpace(user.Username), token)
}

func handleConversationLabelCatalog(w http.ResponseWriter, r *http.Request, username string) {
	response := &apiModels.ConversationLabelsResponse{}
	store, err := getConversationLabelStoreOrError()
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	username = strings.TrimSpace(username)
	if username == "" {
		response.ParseError(fmt.Errorf("server owner is required"))
		RespondInterface(w, response)
		return
	}

	switch r.Method {
	case http.MethodGet:
		activeOnly, err := parseOptionalBoolQuery(r, "active")
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		labels, err := store.FindAllForUser(username, activeOnly)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		response.Labels = labels
		response.ParseSuccess("getting conversation labels")
		RespondSuccess(w, response)
		return

	case http.MethodPost:
		request, err := parseConversationLabelRequest(r)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		active := true
		if request.Active != nil {
			active = *request.Active
		}

		label, err := store.Create(&models.QpConversationLabel{
			User:   username,
			Name:   request.Name,
			Color:  request.Color,
			Active: active,
		})
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		response.Label = label
		response.Affected = 1
		response.ParseSuccess("label created with success")
		RespondSuccess(w, response)
		return

	case http.MethodPut:
		request, err := parseConversationLabelRequest(r)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		if request.ID == 0 {
			response.ParseError(fmt.Errorf("id is required"))
			RespondInterface(w, response)
			return
		}

		current, err := store.FindByIDForUser(request.ID, username)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		active := current.Active
		if request.Active != nil {
			active = *request.Active
		}

		current.Name = request.Name
		current.Color = request.Color
		current.Active = active
		if err := store.Update(current); err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		updated, err := store.FindByIDForUser(request.ID, username)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		response.Label = updated
		response.Affected = 1
		response.ParseSuccess("label updated with success")
		RespondSuccess(w, response)
		return

	case http.MethodDelete:
		id, err := parseConversationLabelID(r)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		if err := store.Delete(id, username); err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		response.Affected = 1
		response.ParseSuccess("label deleted with success")
		RespondSuccess(w, response)
		return
	}

	response.ParseError(fmt.Errorf("method not allowed"))
	RespondInterfaceCode(w, response, http.StatusMethodNotAllowed)
}

func handleConversationChatLabels(w http.ResponseWriter, r *http.Request, username string, serverToken string) {
	response := &apiModels.ConversationLabelsResponse{}
	store, err := getConversationLabelStoreOrError()
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	username = strings.TrimSpace(username)
	serverToken = strings.TrimSpace(serverToken)
	if username == "" {
		response.ParseError(fmt.Errorf("server owner is required"))
		RespondInterface(w, response)
		return
	}
	if serverToken == "" {
		response.ParseError(fmt.Errorf("server token is required"))
		RespondInterface(w, response)
		return
	}

	switch r.Method {
	case http.MethodGet:
		chatID, err := parseConversationChatID(r, nil)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		labels, err := store.FindConversationLabels(serverToken, chatID, username)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		response.ChatID = chatID
		response.Labels = labels
		response.ParseSuccess("getting labels for conversation")
		RespondSuccess(w, response)
		return

	case http.MethodPost, http.MethodDelete:
		request, err := parseConversationChatLabelRequest(r)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		chatID, err := parseConversationChatID(r, request)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		if request.LabelID == 0 {
			response.ParseError(fmt.Errorf("labelid is required"))
			RespondInterface(w, response)
			return
		}

		var affected uint
		if r.Method == http.MethodPost {
			affected, err = store.Assign(serverToken, chatID, request.LabelID, username)
		} else {
			affected, err = store.Remove(serverToken, chatID, request.LabelID, username)
		}
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		labels, err := store.FindConversationLabels(serverToken, chatID, username)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		response.ChatID = chatID
		response.Labels = labels
		response.Affected = affected
		if r.Method == http.MethodPost {
			response.ParseSuccess("label applied with success")
		} else {
			response.ParseSuccess("label removed with success")
		}
		RespondSuccess(w, response)
		return
	}

	response.ParseError(fmt.Errorf("method not allowed"))
	RespondInterfaceCode(w, response, http.StatusMethodNotAllowed)
}

func getConversationLabelServerRecord(r *http.Request) (*models.QpServer, error) {
	if models.WhatsappService == nil || models.WhatsappService.DB == nil || models.WhatsappService.DB.Servers == nil {
		return nil, fmt.Errorf("database service not initialized")
	}

	token := strings.TrimSpace(GetToken(r))
	if token == "" {
		return nil, fmt.Errorf("missing token parameter")
	}

	return models.WhatsappService.DB.Servers.FindByToken(token)
}

func getConversationLabelStoreOrError() (models.QpDataConversationLabelsInterface, error) {
	if models.WhatsappService == nil || models.WhatsappService.DB == nil || models.WhatsappService.DB.ConversationLabels == nil {
		return nil, fmt.Errorf("conversation labels service not initialized")
	}

	return models.WhatsappService.DB.ConversationLabels, nil
}

func parseConversationLabelRequest(r *http.Request) (*conversationLabelRequest, error) {
	request := &conversationLabelRequest{}
	if err := decodeOptionalJSONBody(r, request); err != nil {
		return nil, err
	}

	request.Name = strings.TrimSpace(request.Name)
	request.Color = strings.TrimSpace(request.Color)
	if request.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	return request, nil
}

func parseConversationChatLabelRequest(r *http.Request) (*conversationChatLabelRequest, error) {
	request := &conversationChatLabelRequest{}
	if err := decodeOptionalJSONBody(r, request); err != nil {
		return nil, err
	}

	if request.LabelID == 0 {
		if value := strings.TrimSpace(library.GetRequestParameter(r, "labelid")); value != "" {
			parsed, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid labelid: %w", err)
			}
			request.LabelID = parsed
		}
	}

	return request, nil
}

func parseConversationLabelID(r *http.Request) (int64, error) {
	request := &conversationLabelRequest{}
	if err := decodeOptionalJSONBody(r, request); err != nil {
		return 0, err
	}
	if request.ID != 0 {
		return request.ID, nil
	}

	value := strings.TrimSpace(library.GetRequestParameter(r, "id"))
	if value == "" {
		return 0, fmt.Errorf("id is required")
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id: %w", err)
	}
	return id, nil
}

func parseConversationChatID(r *http.Request, request *conversationChatLabelRequest) (string, error) {
	chatID := strings.TrimSpace(library.GetRequestParameter(r, "chatid"))
	if chatID == "" && request != nil {
		chatID = strings.TrimSpace(request.ChatID)
	}
	if chatID == "" {
		return "", fmt.Errorf("chatid is required")
	}

	formatted, err := whatsapp.FormatEndpoint(chatID)
	if err != nil {
		return "", fmt.Errorf("invalid chatid: %w", err)
	}

	return formatted, nil
}

func parseOptionalBoolQuery(r *http.Request, key string) (*bool, error) {
	value := strings.TrimSpace(library.GetRequestParameter(r, key))
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", key, err)
	}

	return &parsed, nil
}

func decodeOptionalJSONBody(r *http.Request, target interface{}) error {
	if r.Body == nil {
		return nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return nil
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("invalid json body: %w", err)
	}

	return nil
}
