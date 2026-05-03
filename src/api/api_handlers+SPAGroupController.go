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

func getSPAGroupIDParam(r *http.Request) (string, error) {
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

func getSPAOwnedReadyGroupServer(w http.ResponseWriter, r *http.Request) (string, *models.QpWhatsappServer, bool) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return "", nil, false
	}

	token, err := GetSPATokenParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return "", nil, false
	}

	server, err := GetSPAOwnedLiveServer(user, token)
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

// SPAGroupInfoController returns information for a single joined group.
func SPAGroupInfoController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getSPAOwnedReadyGroupServer(w, r)
	if !ok {
		return
	}

	groupID, err := getSPAGroupIDParam(r)
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

// SPAGroupsCreateController creates a new group owned by the current SPA user.
func SPAGroupsCreateController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getSPAOwnedReadyGroupServer(w, r)
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

// SPAGroupLeaveController removes the current live server from a group.
func SPAGroupLeaveController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getSPAOwnedReadyGroupServer(w, r)
	if !ok {
		return
	}

	groupID, err := getSPAGroupIDParam(r)
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

// SPAGroupNameController updates a group subject.
func SPAGroupNameController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getSPAOwnedReadyGroupServer(w, r)
	if !ok {
		return
	}

	groupID, err := getSPAGroupIDParam(r)
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

// SPAGroupDescriptionController updates a group topic.
func SPAGroupDescriptionController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getSPAOwnedReadyGroupServer(w, r)
	if !ok {
		return
	}

	groupID, err := getSPAGroupIDParam(r)
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

// SPAGroupParticipantsController updates members in a group.
func SPAGroupParticipantsController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getSPAOwnedReadyGroupServer(w, r)
	if !ok {
		return
	}

	groupID, err := getSPAGroupIDParam(r)
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

// SPAGroupPhotoController updates or removes a group picture from a remote image URL.
func SPAGroupPhotoController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getSPAOwnedReadyGroupServer(w, r)
	if !ok {
		return
	}

	groupID, err := getSPAGroupIDParam(r)
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

// SPAGroupInviteController returns the invite link for a group.
func SPAGroupInviteController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getSPAOwnedReadyGroupServer(w, r)
	if !ok {
		return
	}

	groupID, err := getSPAGroupIDParam(r)
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

// SPAGroupRevokeInviteController revokes the current invite link and returns the new one.
func SPAGroupRevokeInviteController(w http.ResponseWriter, r *http.Request) {
	_, server, ok := getSPAOwnedReadyGroupServer(w, r)
	if !ok {
		return
	}

	groupID, err := getSPAGroupIDParam(r)
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
