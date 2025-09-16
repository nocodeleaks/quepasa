package api

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
		// Convert webhook to dispatching format
		dispatching := &models.QpDispatching{
			LogStruct:        webhook.LogStruct,
			WhatsappOptions:  webhook.WhatsappOptions,
			ConnectionString: webhook.Url,
			Type:             models.DispatchingTypeWebhook,
			ForwardInternal:  webhook.ForwardInternal,
			TrackId:          webhook.TrackId,
			Extra:            webhook.Extra,
			Failure:          webhook.Failure,
			Success:          webhook.Success,
			Timestamp:        webhook.Timestamp,
			Wid:              webhook.Wid,
		}

		affected, err := server.DispatchingAddOrUpdate(dispatching)
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
		affected, err := server.DispatchingRemove(webhook.Url)
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
		// GET - listar todos os webhooks
		url := r.Header.Get("X-QUEPASA-WHURL")
		logger.Debugf("getting webhooks with filter: '%s'", url)

		dispatching := server.GetDispatchingByFilter(url)
		logger.Debugf("found %d dispatching entries", len(dispatching))

		webhooks := []*models.QpWebhook{}
		for _, item := range dispatching {
			logger.Debugf("checking dispatching item: type=%s, url=%s", item.Type, item.ConnectionString)
			if item.IsWebhook() {
				logger.Debugf("adding webhook: %s", item.ConnectionString)

				// Garantir que Extra seja JSON parseado
				var extraParsed interface{}
				if item.Extra != nil {
					// Se Extra for string (JSON), fazer parse
					if extraStr, ok := item.Extra.(string); ok && extraStr != "" {
						err := json.Unmarshal([]byte(extraStr), &extraParsed)
						if err != nil {
							logger.Warnf("failed to parse extra JSON: %v", err)
							extraParsed = item.Extra
						}
					} else {
						extraParsed = item.Extra
					}
				}

				webhook := &models.QpWebhook{
					LogStruct:       item.LogStruct,
					WhatsappOptions: item.WhatsappOptions,
					Url:             item.ConnectionString,
					ForwardInternal: item.ForwardInternal,
					TrackId:         item.TrackId,
					Extra:           extraParsed,
					Failure:         item.Failure,
					Success:         item.Success,
					Timestamp:       item.Timestamp,
					Wid:             item.Wid,
				}
				webhooks = append(webhooks, webhook)
			}
		}

		logger.Debugf("returning %d webhooks", len(webhooks))
		response.Webhooks = webhooks
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
