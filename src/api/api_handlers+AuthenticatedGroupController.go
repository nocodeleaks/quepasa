package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	apiModels "github.com/nocodeleaks/quepasa/api/models"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

func getGroupIDParam(r *http.Request) (string, error) {
	groupID := strings.TrimSpace(chi.URLParam(r, "groupid"))
	if groupID == "" {
		return "", fmt.Errorf("missing group id parameter")
	}

	decoded, err := url.QueryUnescape(groupID)
	if err == nil && strings.TrimSpace(decoded) != "" {
		groupID = decoded
	}

	if !whatsapp.IsValidGroupId(groupID) {
		return "", fmt.Errorf("invalid group id format: %s", groupID)
	}

	return groupID, nil
}

func getOwnedReadyGroupServer(w http.ResponseWriter, r *http.Request) (string, *models.QpWhatsappServer, bool) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return "", nil, false
	}

	token, err := GetAuthenticatedTokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return "", nil, false
	}

	server, err := GetOwnedLiveServer(user, token)
	if err != nil {
		if err.Error() == "server token not owned by user" {
			RespondErrorCode(w, err, http.StatusForbidden)
			return "", nil, false
		}

		if err.Error() == "server is not active in memory" {
			RespondNotReady(w, err)
			return "", nil, false
		}

		RespondNotFound(w, err)
		return "", nil, false
	}

	return token, server, true
}

// AuthenticatedGroupInfoController returns information for a single joined group.
func AuthenticatedGroupInfoController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getOwnedReadyGroupServer(w, r)
	if !ok {
		return
	}

	groupID, err := getGroupIDParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	group, err := server.GetGroupManager().GetGroupInfo(groupID)
	if err != nil {
		RespondServerError(server, w, err)
		return
	}

	response := &apiModels.SingleGroupResponse{}
	response.GroupInfo = group
	RespondSuccess(w, response)
}

// AuthenticatedGroupCreateController creates a new group owned by the current authenticated user.
func AuthenticatedGroupCreateController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getOwnedReadyGroupServer(w, r)
	if !ok {
		return
	}

	var request struct {
		Title        string   `json:"title"`
		Participants []string `json:"participants"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		RespondErrorCode(w, fmt.Errorf("invalid JSON body: %w", err), http.StatusBadRequest)
		return
	}

	request.Title = strings.TrimSpace(request.Title)
	if request.Title == "" {
		RespondErrorCode(w, fmt.Errorf("title is required"), http.StatusBadRequest)
		return
	}

	if len(request.Title) > 25 {
		RespondErrorCode(w, fmt.Errorf("group title is limited to 25 characters"), http.StatusBadRequest)
		return
	}

	if len(request.Participants) == 0 {
		RespondErrorCode(w, fmt.Errorf("participants are required"), http.StatusBadRequest)
		return
	}

	groupInfo, err := server.GetGroupManager().CreateGroupExtendedWithOptions(map[string]interface{}{
		"title":        request.Title,
		"participants": whatsapp.PhonesToWids(request.Participants),
	})
	if err != nil {
		RespondServerError(server, w, err)
		return
	}

	response := &apiModels.SingleGroupResponse{}
	response.GroupInfo = groupInfo
	RespondSuccess(w, response)
}

// AuthenticatedGroupLeaveController removes the current live server from a group.
func AuthenticatedGroupLeaveController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getOwnedReadyGroupServer(w, r)
	if !ok {
		return
	}

	groupID, err := getGroupIDParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	if err := server.GetGroupManager().LeaveGroup(groupID); err != nil {
		RespondServerError(server, w, err)
		return
	}

	RespondSuccess(w, map[string]interface{}{"result": "successfully left the group"})
}

// AuthenticatedGroupNameController updates a group subject.
func AuthenticatedGroupNameController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getOwnedReadyGroupServer(w, r)
	if !ok {
		return
	}

	groupID, err := getGroupIDParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	var request struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		RespondErrorCode(w, fmt.Errorf("invalid JSON body: %w", err), http.StatusBadRequest)
		return
	}

	request.Name = strings.TrimSpace(request.Name)
	if request.Name == "" {
		RespondErrorCode(w, fmt.Errorf("name is required"), http.StatusBadRequest)
		return
	}

	groupInfo, err := server.GetGroupManager().UpdateGroupSubject(groupID, request.Name)
	if err != nil {
		RespondServerError(server, w, err)
		return
	}

	response := &apiModels.SingleGroupResponse{}
	response.GroupInfo = groupInfo
	RespondSuccess(w, response)
}

// AuthenticatedGroupDescriptionController updates a group topic.
func AuthenticatedGroupDescriptionController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getOwnedReadyGroupServer(w, r)
	if !ok {
		return
	}

	groupID, err := getGroupIDParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	var request struct {
		Topic string `json:"topic"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		RespondErrorCode(w, fmt.Errorf("invalid JSON body: %w", err), http.StatusBadRequest)
		return
	}

	groupInfo, err := server.GetGroupManager().UpdateGroupTopic(groupID, request.Topic)
	if err != nil {
		RespondServerError(server, w, err)
		return
	}

	response := &apiModels.SingleGroupResponse{}
	response.GroupInfo = groupInfo
	RespondSuccess(w, response)
}

// AuthenticatedGroupParticipantsController updates members in a group.
func AuthenticatedGroupParticipantsController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getOwnedReadyGroupServer(w, r)
	if !ok {
		return
	}

	groupID, err := getGroupIDParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	var request struct {
		Participants []string `json:"participants"`
		Action       string   `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		RespondErrorCode(w, fmt.Errorf("invalid JSON body: %w", err), http.StatusBadRequest)
		return
	}

	if len(request.Participants) == 0 {
		RespondErrorCode(w, fmt.Errorf("at least one participant is required"), http.StatusBadRequest)
		return
	}

	action := strings.ToLower(strings.TrimSpace(request.Action))
	validActions := map[string]bool{"add": true, "remove": true, "promote": true, "demote": true}
	if !validActions[action] {
		RespondErrorCode(w, fmt.Errorf("invalid action, must be one of: add, remove, promote, demote"), http.StatusBadRequest)
		return
	}

	result, err := server.GetGroupManager().UpdateGroupParticipants(groupID, whatsapp.PhonesToWids(request.Participants), action)
	if err != nil {
		RespondServerError(server, w, err)
		return
	}

	response := &apiModels.ParticipantResponse{}
	response.Total = len(result)
	response.Participants = result
	response.ParseSuccess(fmt.Sprintf("Group participants updated successfully. Action: %s", action))
	RespondSuccess(w, response)
}

// AuthenticatedGroupPhotoController updates or removes a group picture from a remote image URL.
func AuthenticatedGroupPhotoController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getOwnedReadyGroupServer(w, r)
	if !ok {
		return
	}

	groupID, err := getGroupIDParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	var request struct {
		ImageURL  string `json:"image_url"`
		RemoveImg bool   `json:"remove_img"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		RespondErrorCode(w, fmt.Errorf("invalid JSON body: %w", err), http.StatusBadRequest)
		return
	}

	if request.RemoveImg {
		if _, err := server.GetGroupManager().UpdateGroupPhoto(groupID, nil); err != nil {
			RespondServerError(server, w, err)
			return
		}

		RespondSuccess(w, map[string]interface{}{"result": "Group photo removed successfully"})
		return
	}

	request.ImageURL = strings.TrimSpace(request.ImageURL)
	if request.ImageURL == "" {
		RespondErrorCode(w, fmt.Errorf("image_url is required"), http.StatusBadRequest)
		return
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Get(request.ImageURL)
	if err != nil {
		RespondErrorCode(w, fmt.Errorf("failed to download image: %w", err), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		RespondErrorCode(w, fmt.Errorf("failed to download image, server returned status: %s", resp.Status), http.StatusBadRequest)
		return
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		RespondErrorCode(w, fmt.Errorf("failed to read image data: %w", err), http.StatusBadRequest)
		return
	}

	if len(imageData) < 10000 || len(imageData) > 500000 {
		RespondErrorCode(w, fmt.Errorf("image size not ideal for WhatsApp (should be between 10KB and 500KB)"), http.StatusBadRequest)
		return
	}

	pictureID, err := server.GetGroupManager().UpdateGroupPhoto(groupID, imageData)
	if err != nil {
		RespondServerError(server, w, err)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"result":    "Group photo updated successfully",
		"pictureId": pictureID,
	})
}

// AuthenticatedGroupInviteController returns the invite link for a group.
func AuthenticatedGroupInviteController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getOwnedReadyGroupServer(w, r)
	if !ok {
		return
	}

	groupID, err := getGroupIDParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	url, err := server.GetGroupManager().GetInvite(groupID)
	if err != nil {
		RespondServerError(server, w, err)
		return
	}

	response := &apiModels.InviteResponse{}
	response.Url = url
	RespondSuccess(w, response)
}

// AuthenticatedGroupRevokeInviteController revokes the current invite link and returns the new one.
func AuthenticatedGroupRevokeInviteController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getOwnedReadyGroupServer(w, r)
	if !ok {
		return
	}

	groupID, err := getGroupIDParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	url, err := server.GetGroupManager().RevokeInvite(groupID)
	if err != nil {
		RespondServerError(server, w, err)
		return
	}

	response := &apiModels.InviteResponse{}
	response.Url = url
	RespondSuccess(w, response)
}
