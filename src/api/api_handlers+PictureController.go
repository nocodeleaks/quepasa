package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// InformationController renders route GET "/{version}/bot/{token}"
func PictureController(w http.ResponseWriter, r *http.Request) {

	response := &models.QpPictureResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Checking for ready state
	status := server.GetStatus()
	if status != whatsapp.Ready {
		err = &ApiServerNotReadyException{Wid: server.GetWId(), Status: status}
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
		return
	}

	// Getting ChatId parameter
	chatId := library.GetChatId(r)
	pictureId := GetPictureId(r)

	switch os := r.Method; os {
	case http.MethodPost:
		// Declare a new Person struct.
		var p whatsapp.WhatsappProfilePicture

		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err = json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			jsonErr := fmt.Errorf("invalid json body: %s", err.Error())
			response.ParseError(jsonErr)
			RespondInterface(w, response)
			return
		}

		if len(p.Id) > 0 {
			pictureId = p.Id
		}

		if len(p.ChatId) > 0 {
			chatId = p.ChatId
		}
	}

	chatId, err = whatsapp.FormatEndpoint(chatId)
	if err != nil {
		chatIdErr := fmt.Errorf("invalid chatId: %s", err.Error())
		response.ParseError(chatIdErr)
		RespondInterface(w, response)
		return
	}

	info, err := server.GetProfilePicture(chatId, pictureId)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	if info != nil {
		response.Info = info
		if strings.Contains(r.URL.Path, "picdata") {
			resp, err := http.Get(response.Info.Url)
			if err != nil {
				response.ParseError(err)
				RespondInterface(w, response)
				return
			}
			defer resp.Body.Close()

			content, err := io.ReadAll(resp.Body)
			if err != nil {
				response.ParseError(err)
				RespondInterface(w, response)
				return
			}

			w.Header().Set("Content-Disposition", "attachment; filename="+response.Info.Id+".jpg")
			w.WriteHeader(http.StatusOK)
			w.Write(content)
			return
		} else {
			RespondSuccess(w, response)
			return
		}
	} else {
		response.ParseSuccess("not modified")
		RespondInterfaceCode(w, response, 304)
		return
	}
}
