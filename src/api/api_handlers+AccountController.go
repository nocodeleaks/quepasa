package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/nbutton23/zxcvbn-go"
	models "github.com/nocodeleaks/quepasa/models"
)

//region CONTROLLER - HEALTH

// AccountController manages user accounts and authentication
// @Summary Manage user accounts
// @Description Create, update, or manage user accounts (master access required)
// @Tags Application
// @Accept json
// @Produce json
// @Param request body object{username=string,password=string} false "Account request"
// @Success 200 {object} models.QpResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /account [post]
func AccountController(w http.ResponseWriter, r *http.Request) {

	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpResponse{}

	master := IsMatchForMaster(r)
	if !master {
		response.ParseError(errors.New("no puedo creer que seas tan caradura"))
		RespondInterface(w, response)
		return
	}

	switch os := r.Method; os {
	case http.MethodPost:

		// reading body to avoid converting to json if empty
		body, err := io.ReadAll(r.Body)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		// Declare a new Person struct.
		var request *models.QpAccountUpdateRequest

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

		if len(request.Password) > 0 {
			err := PasswordUpdate(request.User, request.Password)
			if err != nil {
				jsonError := fmt.Errorf("error on password update: %v", err.Error())
				response.ParseError(jsonError)
				RespondInterface(w, response)
				return
			}

			response.ParseSuccess("password updated")
			return
		}

		response.ParseError(errors.New("no action acknoledge"))
		RespondInterface(w, response)
		return
	default:
		response.ParseError(errors.New("nothing to see here for now"))
		RespondInterface(w, response)
		return
	}
}

//endregion

func PasswordUpdate(user string, password string) error {
	if len(user) == 0 {
		return errors.New("missing user parameter")
	}

	res := zxcvbn.PasswordStrength(password, nil)
	if res.Score < 1 {
		return errors.New("password is too weak")
	}

	exists, err := models.WhatsappService.DB.Users.Exists(user)
	if err != nil {
		return fmt.Errorf("error on database check if user exists: %s", err.Error())
	}

	if !exists {
		return fmt.Errorf("user not found: %s", user)
	}

	err = models.WhatsappService.DB.Users.UpdatePassword(user, password)
	if err != nil {
		return fmt.Errorf("error on database updating password: %s", err.Error())
	}

	return nil
}
