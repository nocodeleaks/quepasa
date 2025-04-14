package controllers

import (
	"fmt"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
)

//region CONTROLLER - HEALTH

func HealthController(w http.ResponseWriter, r *http.Request) {

	// setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpHealthResponse{}

	master := IsMatchForMaster(r)
	if master {
		response.Items = models.WhatsappService.GetHealth()
		response.Success = All(response.Items, models.QpHealthResponseItem.GetHealth)
		RespondInterface(w, response)
		return
	} else {
		server, err := GetServer(r)
		if err != nil {
			response.ParseError(err)
			RespondInterface(w, response)
			return
		}

		status := server.GetStatus()
		response.Success = status.IsHealthy()
		response.Status = fmt.Sprintf("server status is %s", status.String())
		RespondInterface(w, response)
		return
	}
}

//endregion

func All[T any](ts []T, pred func(T) bool) bool {
	for _, t := range ts {
		if !pred(t) {
			return false
		}
	}
	return true
}
