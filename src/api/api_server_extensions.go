package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	environment "github.com/nocodeleaks/quepasa/environment"
	models "github.com/nocodeleaks/quepasa/models"
)

/*
<summary>

	Find a whatsapp server by token passed on Url Path parameters

</summary>
*/
func GetServer(r *http.Request) (server *models.QpWhatsappServer, err error) {
	token := GetToken(r)
	return models.GetServerFromToken(token)
}

// <summary>Find a whatsapp server by token passed on Url Path parameters</summary>
func GetServerRespondOnError(w http.ResponseWriter, r *http.Request) (server *models.QpWhatsappServer, err error) {
	token := GetToken(r)
	server, err = models.GetServerFromToken(token)
	if err != nil {
		RespondNoContentV2(w, fmt.Errorf("token '%s' not found", token))
	}
	return
}

func GetServerFromMaster(r *http.Request) (server *models.QpWhatsappServer, err error) {
	system := environment.Settings.API.MasterKey
	if len(system) == 0 {
		return nil, errors.New("server is not allowed to use this method")
	}

	request := GetMasterKey(r)
	if !strings.EqualFold(system, request) {
		return nil, errors.New("dont even try to trick me, first strike")
	}

	return models.GetServerFirstAvailable()
}

// <summary>Checks if was passed a valid master key</summary>
func IsMatchForMaster(r *http.Request) bool {
	system := environment.Settings.API.MasterKey
	if len(system) > 0 {
		request := GetMasterKey(r)
		if strings.EqualFold(system, request) {
			return true
		}
	}

	return false
}
