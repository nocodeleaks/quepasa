package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	apiModels "github.com/nocodeleaks/quepasa/api/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

func getChatIDParam(r *http.Request) (string, error) {
	chatID := strings.TrimSpace(chi.URLParam(r, "chatid"))
	if chatID == "" {
		return "", fmt.Errorf("missing chat id parameter")
	}

	decoded, err := url.QueryUnescape(chatID)
	if err == nil && strings.TrimSpace(decoded) != "" {
		chatID = decoded
	}

	formatted, err := whatsapp.FormatEndpoint(chatID)
	if err != nil {
		return "", fmt.Errorf("invalid chatId: %s", err.Error())
	}

	return formatted, nil
}

// AuthenticatedPictureInfoController returns profile picture metadata for a chat using authenticated API access.
func AuthenticatedPictureInfoController(w http.ResponseWriter, r *http.Request) {
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
		if err.Error() == "server token not owned by user" {
			RespondErrorCode(w, err, http.StatusForbidden)
			return
		}
		if err.Error() == "server is not active in memory" {
			RespondNotReady(w, err)
			return
		}
		RespondNotFound(w, err)
		return
	}

	if err := EnsureLiveServerReady(server); err != nil {
		if _, ok := err.(*ApiServerNotReadyException); ok {
			RespondNotReady(w, err)
			return
		}
		RespondServerError(server, w, err)
		return
	}

	chatID, err := getChatIDParam(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}

	pictureID := strings.TrimSpace(chi.URLParam(r, "pictureid"))
	if pictureID != "" {
		if decoded, err := url.QueryUnescape(pictureID); err == nil && strings.TrimSpace(decoded) != "" {
			pictureID = decoded
		}
	}

	info, err := server.GetProfilePicture(chatID, pictureID)
	if err != nil {
		RespondServerError(server, w, err)
		return
	}

	response := &apiModels.PictureResponse{}
	if info == nil {
		response.ParseSuccess("not modified")
		RespondInterfaceCode(w, response, http.StatusNotModified)
		return
	}

	response.Info = info
	RespondSuccess(w, response)
}
