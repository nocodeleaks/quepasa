package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
)

//region CONTROLLER - WEBHOOK

func WebhookController(w http.ResponseWriter, r *http.Request) {

	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpWebhookResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	logger := server.GetLogger()

	// reading body to avoid converting to json if empty
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Declare a new Person struct.
	var webhook *models.QpWebhook

	if len(body) > 0 {

		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err = json.Unmarshal(body, &webhook)
		if err != nil {
			jsonError := fmt.Errorf("error converting body to json: %v", err.Error())
			response.ParseError(jsonError)
			RespondInterface(w, response)
			return
		}
	}

	// creating an empty webhook, to filter or clear it all
	if webhook == nil {
		webhook = &models.QpWebhook{}
	}

	// updating wid for logging and response headers
	webhook.Wid = server.Wid

	switch os := r.Method; os {
	case http.MethodPost:
		affected, err := server.WebhookAddOrUpdate(webhook)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
		} else {
			response.Affected = affected
			response.ParseSuccess("updated with success")
			RespondSuccess(w, response)
			if affected > 0 {
				logger.Infof("updating webhook url=%s, items affected: %v", webhook.Url, affected)
			}
		}
		return
	case http.MethodDelete:
		affected, err := server.WebhookRemove(webhook.Url)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
		} else {
			response.Affected = affected
			response.ParseSuccess("deleted with success")
			RespondSuccess(w, response)
			if affected > 0 {
				logger.Infof("removing webhook url=%s, items affected: %v", webhook.Url, affected)
			}
		}
		return
	default:
		url := r.Header.Get("X-QUEPASA-WHURL")
		response.Webhooks = server.GetWebHooksByUrl(url)
		if len(url) > 0 {
			response.ParseSuccess(fmt.Sprintf("getting with filter, url=%s", url))
		} else {
			response.ParseSuccess("getting without filter")
		}

		RespondSuccess(w, response)
		return
	}
}

//endregion
