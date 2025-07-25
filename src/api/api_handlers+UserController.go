package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
)

//region CONTROLLER - User

func UserController(w http.ResponseWriter, r *http.Request) {

	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpInfoResponse{}

	// reading body to avoid converting to json if empty
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Declare a new Person struct.
	var request *models.QpUser

	if len(body) > 0 {

		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err = json.Unmarshal(body, &request)
		if err != nil {
			jsonError := fmt.Errorf("error converting body to json: %v", err.Error())
			response.ParseError(jsonError)
			RespondInterface(w, response)
			return
		}
	}

	if request == nil || len(request.Username) == 0 {
		jsonErr := fmt.Errorf("invalid user body: %s", string(body))
		response.ParseError(jsonErr)
		RespondInterface(w, response)
		return
	}

	// searching user
	request, err = models.WhatsappService.DB.Users.Find(request.Username)
	if err != nil {
		jsonError := fmt.Errorf("user not found: %v", err.Error())
		response.ParseError(jsonError)
		RespondInterface(w, response)
		return
	}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	server.User = request.Username
	err = server.Save("updating username")
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	response.PatchSuccess(server, "server attached for user: "+request.Username)
	RespondSuccess(w, response)
}

//endregion
