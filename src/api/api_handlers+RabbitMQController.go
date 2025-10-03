package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	models "github.com/nocodeleaks/quepasa/models"
)

//region CONTROLLER - RABBITMQ

// RabbitMQController manages RabbitMQ configurations
// @Summary Manage RabbitMQ configurations
// @Description Create, get, or delete RabbitMQ configurations for message queueing
// @Tags RabbitMQ
// @Accept json
// @Produce json
// @Param request body object{connection_string=string,exchange=string,routing_key=string} false "RabbitMQ config (for POST)"
// @Param connection_string query string false "Connection string (for DELETE)"
// @Success 200 {object} models.QpRabbitMQResponse
// @Failure 400 {object} models.QpResponse
// @Security ApiKeyAuth
// @Router /rabbitmq [get]
// @Router /rabbitmq [post]
// @Router /rabbitmq [delete]
func RabbitMQController(w http.ResponseWriter, r *http.Request) {

	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpRabbitMQResponse{}

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

	// Declare a new RabbitMQ config struct.
	var rabbitmqConfig *models.QpRabbitMQConfig

	if len(body) > 0 {

		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err = json.Unmarshal(body, &rabbitmqConfig)
		if err != nil {
			jsonError := fmt.Errorf("error converting body to json: %v", err.Error())
			response.ParseError(jsonError)
			RespondInterface(w, response)
			return
		}
	}

	// creating an empty rabbitmq config, to filter or clear it all
	if rabbitmqConfig == nil {
		rabbitmqConfig = &models.QpRabbitMQConfig{}
	}

	// updating wid for logging and response headers
	rabbitmqConfig.Wid = server.Wid

	switch os := r.Method; os {
	case http.MethodPost:
		// Only connection_string is required now
		// exchange_name and routing_key are fixed for all bots
		if rabbitmqConfig.ConnectionString == "" {
			response.ParseError(fmt.Errorf("connection_string is required"))
			RespondInterface(w, response)
			return
		}

		// Validate RabbitMQ configuration
		err = rabbitmqConfig.ValidateConfig()
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		// Convert RabbitMQ config to dispatching
		dispatching := rabbitmqConfig.ToDispatching()
		affected, err := server.DispatchingAddOrUpdate(dispatching)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
		} else {
			response.Affected = affected
			response.ParseSuccess("updated with success")
			RespondSuccess(w, response)
			if affected > 0 {
				logger.Infof("updating rabbitmq connection=%s (using fixed QuePasa exchange), items affected: %v", rabbitmqConfig.ConnectionString, affected)
			}
		}
		return
	case http.MethodDelete:
		// Para DELETE, precisamos apenas do connection_string
		var connectionString string

		// Tentar obter da query string primeiro
		connectionString = r.URL.Query().Get("connection_string")

		// Se não veio da query string, tentar do body
		if connectionString == "" && rabbitmqConfig != nil {
			connectionString = rabbitmqConfig.ConnectionString
		}

		// Se ainda não temos connection_string, retornar erro
		if connectionString == "" {
			response.ParseError(fmt.Errorf("connection_string is required for delete operation"))
			RespondInterface(w, response)
			return
		}

		// Try multiple variations of the connection string to handle URL encoding issues
		connectionStrings := []string{
			connectionString,                  // Original
			url.QueryEscape(connectionString), // URL encoded
		}

		// Add URL decoded version if different
		if decoded, err := url.QueryUnescape(connectionString); err == nil && decoded != connectionString {
			connectionStrings = append(connectionStrings, decoded)
		}

		// Add manual fix for common encoding issue
		if len(connectionString) >= 2 && connectionString[len(connectionString)-2:] == "//" {
			// Replace "//" with "/%2F" - common encoding difference
			fixedConnectionString := connectionString[:len(connectionString)-2] + "/%2F"
			connectionStrings = append(connectionStrings, fixedConnectionString)
		}

		// Try to remove with each variation
		var totalAffected uint = 0
		var lastErr error

		for i, cs := range connectionStrings {
			if cs == connectionString && i > 0 {
				continue // Skip if it's the same as original (already tried)
			}

			affected, err := server.DispatchingRemove(cs)
			if err != nil {
				lastErr = err
			} else {
				totalAffected += affected
				if affected > 0 {
					break // Success, stop trying other variations
				}
			}
		}

		if totalAffected > 0 {
			response.Affected = totalAffected
			response.ParseSuccess("deleted with success")
			RespondSuccess(w, response)
			logger.Infof("removing rabbitmq connection_string=%s, items affected: %v", connectionString, totalAffected)
		} else {
			if lastErr != nil {
				response.ParseError(lastErr)
			} else {
				response.ParseError(fmt.Errorf("no matching connection string found for deletion"))
			}
			RespondInterface(w, response)
		}
		return
	default:
		// GET - listar todas as configurações RabbitMQ
		response.RabbitMQ = server.GetRabbitMQConfigs()
		response.ParseSuccess("getting all RabbitMQ configurations (all use fixed QuePasa Exchange)")
		RespondSuccess(w, response)
		return
	}
}

//endregion
