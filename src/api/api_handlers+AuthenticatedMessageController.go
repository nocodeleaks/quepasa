package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	apiModels "github.com/nocodeleaks/quepasa/api/models"
	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// AuthenticatedServerSendController sends a message through the current send request model
// while enforcing SPA JWT auth and server ownership checks.
func AuthenticatedServerSendController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	server, err := GetOwnedLiveServer(user, token)
	if err != nil {
		respondSPASessionLookupError(w, err)
		return
	}

	SendAnyWithServer(w, r, server)
}

// AuthenticatedServerArchiveChatController archives or unarchives a chat through the SPA auth surface.
func AuthenticatedServerArchiveChatController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	server, err := GetOwnedLiveServer(user, token)
	if err != nil {
		respondSPASessionLookupError(w, err)
		return
	}

	ArchiveChatWithServer(w, r, server)
}

// AuthenticatedServerPresenceController sends typing/presence updates through the SPA auth surface.
func AuthenticatedServerPresenceController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	server, err := GetOwnedLiveServer(user, token)
	if err != nil {
		respondSPASessionLookupError(w, err)
		return
	}

	ChatPresenceWithServer(w, r, server)
}

// AuthenticatedWebHooksController exposes webhook CRUD through the SPA auth surface.
func AuthenticatedWebHooksController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	server, err := GetOwnedLiveServer(user, token)
	if err != nil {
		respondSPASessionLookupError(w, err)
		return
	}

	WebhookWithServer(w, r, server)
}

// AuthenticatedRabbitMQController exposes RabbitMQ CRUD through the SPA auth surface.
func AuthenticatedRabbitMQController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	server, err := GetOwnedLiveServer(user, token)
	if err != nil {
		respondSPASessionLookupError(w, err)
		return
	}

	RabbitMQWithServer(w, r, server)
}

// AuthenticatedServerMessagesController returns paginated messages for a live server owned by the user.
func AuthenticatedServerMessagesController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	server, err := GetOwnedLiveServer(user, token)
	if err != nil {
		respondSPASessionLookupError(w, err)
		return
	}

	if err := EnsureLiveServerReady(server); err != nil {
		respondSPASessionReadyError(w, err)
		return
	}

	page := 1
	limit := 50
	if pageStr := strings.TrimSpace(r.URL.Query().Get("page")); pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}
	if limitStr := strings.TrimSpace(r.URL.Query().Get("limit")); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 500 {
				limit = 500
			}
		}
	}

	timestamp, err := GetTimestamp(r)
	if err != nil {
		response := &apiModels.ReceiveResponse{}
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	filters := GetReceiveMessageFilters(r)
	messages := GetOrderedMessagesWithFilters(server, timestamp, filters)
	total := len(messages)
	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}
	if totalPages > 0 && page > totalPages {
		page = totalPages
	}

	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	response := &apiModels.ReceiveResponse{}
	response.Server = server.QpServer
	response.Total = uint64(total)
	response.Page = page
	response.Limit = limit
	response.TotalPages = totalPages
	if start < end {
		response.Messages = messages[start:end]
	} else {
		response.Messages = []whatsapp.WhatsappMessage{}
	}

	msg := buildSPAMessageResponseSummary(timestamp, filters, page, totalPages, len(response.Messages))
	response.ParseSuccess(msg)
	RespondSuccess(w, response)
}

// AuthenticatedServerEditMessageController edits the content of a cached message for a live server.
func AuthenticatedServerEditMessageController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	messageID := GetMessageId(r)
	if strings.TrimSpace(messageID) == "" {
		RespondErrorCode(w, fmt.Errorf("missing messageid parameter"), http.StatusBadRequest)
		return
	}

	server, err := GetOwnedLiveServer(user, token)
	if err != nil {
		respondSPASessionLookupError(w, err)
		return
	}

	if err := EnsureLiveServerReady(server); err != nil {
		respondSPASessionReadyError(w, err)
		return
	}

	var request struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		RespondErrorCode(w, fmt.Errorf("invalid json body: %s", err.Error()), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(request.Content) == "" {
		RespondErrorCode(w, fmt.Errorf("content is required"), http.StatusBadRequest)
		return
	}

	if err := server.Edit(messageID, request.Content); err != nil {
		RespondServerError(server, w, err)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"result":    "success",
		"messageId": messageID,
	})
}

// AuthenticatedServerRevokeMessageController revokes a cached message for a live server.
func AuthenticatedServerRevokeMessageController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	messageID := GetMessageId(r)
	if strings.TrimSpace(messageID) == "" {
		RespondErrorCode(w, fmt.Errorf("missing messageid parameter"), http.StatusBadRequest)
		return
	}

	server, err := GetOwnedLiveServer(user, token)
	if err != nil {
		respondSPASessionLookupError(w, err)
		return
	}

	if err := EnsureLiveServerReady(server); err != nil {
		respondSPASessionReadyError(w, err)
		return
	}

	if err := server.Revoke(messageID); err != nil {
		RespondServerError(server, w, err)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"result":    "success",
		"messageId": messageID,
	})
}

// AuthenticatedServerDownloadMediaController downloads message media while preserving current legacy defaults.
func AuthenticatedServerDownloadMediaController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	messageID := GetMessageId(r)
	if strings.TrimSpace(messageID) == "" {
		RespondErrorCode(w, fmt.Errorf("missing messageid parameter"), http.StatusBadRequest)
		return
	}

	server, err := GetOwnedLiveServer(user, token)
	if err != nil {
		respondSPASessionLookupError(w, err)
		return
	}

	if err := EnsureLiveServerReady(server); err != nil {
		respondSPASessionReadyError(w, err)
		return
	}

	attachment, err := server.Download(messageID, GetCache(r))
	if err != nil {
		RespondServerError(server, w, err)
		return
	}

	filename := attachment.FileName
	if filename == "" {
		if exten, ok := library.TryGetExtensionFromMimeType(attachment.Mimetype); ok {
			filename = messageID + exten
		}
	}

	dispositionType := "attachment"
	if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("disposition")), "inline") {
		dispositionType = "inline"
	}
	if filename != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("%s; filename=%q", dispositionType, filename))
	} else {
		w.Header().Set("Content-Disposition", dispositionType)
	}

	if attachment.Mimetype != "" {
		w.Header().Set("Content-Type", attachment.Mimetype)
	}

	content := attachment.GetContent()
	if content == nil {
		RespondErrorCode(w, fmt.Errorf("download returned empty content"), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(*content)
}

func respondSPASessionLookupError(w http.ResponseWriter, err error) {
	switch err.Error() {
	case "server token not owned by user":
		RespondErrorCode(w, err, http.StatusForbidden)
	case "server is not active in memory":
		RespondNotReady(w, err)
	default:
		RespondNotFound(w, err)
	}
}

func respondSPASessionReadyError(w http.ResponseWriter, err error) {
	response := &models.QpResponse{}
	response.ParseError(err)
	RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
}

// Backward compatibility aliases for server-named functions
func respondServerLookupError(w http.ResponseWriter, err error) {
	respondSPASessionLookupError(w, err)
}

func respondSPAServerReadyError(w http.ResponseWriter, err error) {
	respondSPASessionReadyError(w, err)
}

func buildSPAMessageResponseSummary(timestamp int64, filters ReceiveMessageFilters, page, totalPages, count int) string {
	var parts []string
	if timestamp > 0 {
		parts = append(parts, fmt.Sprintf("getting with timestamp: %v => %s", timestamp, time.Unix(timestamp, 0)))
	} else {
		parts = append(parts, "getting without timestamp filter")
	}

	if filters.Exceptions != "" {
		parts = append(parts, "exceptions="+filters.Exceptions)
	}
	if filters.Type != "" {
		parts = append(parts, "type="+filters.Type)
	}
	if filters.Category != "" {
		parts = append(parts, "category="+filters.Category)
	}
	if filters.Search != "" {
		parts = append(parts, "search="+filters.Search)
	}
	if filters.FromMe != "" {
		parts = append(parts, "fromme="+filters.FromMe)
	}
	if filters.FromHistory != "" {
		parts = append(parts, "fromhistory="+filters.FromHistory)
	}
	if filters.ChatID != "" {
		parts = append(parts, "chatid="+filters.ChatID)
	}
	if filters.MessageID != "" {
		parts = append(parts, "messageid="+filters.MessageID)
	}
	if filters.TrackID != "" {
		parts = append(parts, "trackid="+filters.TrackID)
	}

	displayTotalPages := totalPages
	if displayTotalPages == 0 {
		displayTotalPages = 1
	}

	parts = append(parts, fmt.Sprintf("page %d/%d (%d messages)", page, displayTotalPages, count))
	return strings.Join(parts, ", ")
}
