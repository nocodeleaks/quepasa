package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	legacy "github.com/nocodeleaks/quepasa/api/legacy"

	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	runtime "github.com/nocodeleaks/quepasa/runtime"
)

// CurrentAPIVersion is the latest versioned alias exposed by the legacy HTTP API.
const CurrentAPIVersion string = "v4"

func legacyAliases(includeUnversioned bool) []string {
	aliases := []string{"/current", "/" + CurrentAPIVersion}
	if includeUnversioned {
		aliases = append([]string{""}, aliases...)
	}
	return aliases
}

// RegisterAPIControllers wires the legacy/public HTTP API.
//
// The route set is intentionally registered under multiple aliases to preserve
// compatibility with existing clients while newer surfaces are introduced.
func RegisterAPIControllers(r chi.Router, includeUnversioned bool) {
	legacy.RegisterAPIControllers(r, legacy.Config{CurrentAPIVersion: CurrentAPIVersion, Aliases: legacyAliases(includeUnversioned)}, legacyHandlers())
}

// CommandController manages bot server commands.
//
//	@Summary		Execute bot commands
//	@Description	Execute control commands for the bot server (start, stop, restart)
//	@Tags			Bot
//	@Accept			json
//	@Produce		json
//	@Param			action	query		string	true	"Command action"	Enums(start, stop, restart)
//	@Success		200		{object}	models.QpResponse
//	@Failure		400		{object}	models.QpResponse
//	@Security		ApiKeyAuth
//	@Router			/command [get]
func CommandController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpResponse{}

	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	action := library.GetRequestParameter(r, "action")
	switch action {
	case "start":
		err = runtime.StartSession(server)
		if err == nil {
			response.ParseSuccess("started")
		}
	case "stop":
		err = runtime.StopSession(server, "command")
		if err == nil {
			response.ParseSuccess("stopped")
		}
	case "restart":
		err = runtime.RestartSession(server)
		if err == nil {
			response.ParseSuccess("restarted")
		}
	case "status":
		err = fmt.Errorf("status command has been removed, please use /health endpoint instead")
	case "groups":
		// These toggles remain part of the current command surface until the
		// authenticated browser-oriented surface fully owns configuration through explicit endpoints.
		_, err := runtime.ToggleSessionOption(server, "groups")
		if err == nil {
			message := "groups toggled: " + server.Groups.String()
			response.ParseSuccess(message)
		}
	case "broadcasts":
		_, err := runtime.ToggleSessionOption(server, "broadcasts")
		if err == nil {
			message := "broadcasts toggled: " + server.Broadcasts.String()
			response.ParseSuccess(message)
		}
	case "readreceipts":
		_, err := runtime.ToggleSessionOption(server, "readreceipts")
		if err == nil {
			message := "readreceipts toggled: " + server.ReadReceipts.String()
			response.ParseSuccess(message)
		}
	case "calls":
		_, err := runtime.ToggleSessionOption(server, "calls")
		if err == nil {
			message := "calls toggled: " + server.Calls.String()
			response.ParseSuccess(message)
		}
	case "debug":
		_, err := runtime.ToggleSessionDebug(server)
		if err == nil {
			message := "debug toggled: " + fmt.Sprintf("%t", server.Devel)
			response.ParseSuccess(message)
		}
	default:
		err = fmt.Errorf("invalid action: {%s}, try {start,stop,restart,groups,broadcasts,readreceipts,calls,debug}", action)
	}

	if err != nil {
		response.ParseError(err)
	}

	RespondInterface(w, response)
}
