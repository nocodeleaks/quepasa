package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	environment "github.com/nocodeleaks/quepasa/environment"
	models "github.com/nocodeleaks/quepasa/models"
)

// dispatchTypesToCSV renders the env DispatchTypes set as a sorted CSV string for display.
func dispatchTypesToCSV(types map[string]bool) string {
	keys := make([]string, 0, len(types))
	for k, enabled := range types {
		if enabled {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

// CanonicalSettingsGetController returns env defaults + current global overrides. Master key required.
func CanonicalSettingsGetController(w http.ResponseWriter, r *http.Request) {
	if !IsMatchForMaster(r) {
		RespondErrorCode(w, fmt.Errorf("master key required"), http.StatusUnauthorized)
		return
	}

	env := environment.Settings.Messages
	RespondSuccess(w, map[string]interface{}{
		"env": map[string]interface{}{
			"store_retention_days": env.RetentionDays,
			"dispatch_types":       dispatchTypesToCSV(env.DispatchTypes),
		},
		"global": models.GetGlobalMessageConfig(),
	})
}

// CanonicalSettingsPutController writes the global overrides (runtime, no restart). Master key required.
func CanonicalSettingsPutController(w http.ResponseWriter, r *http.Request) {
	if !IsMatchForMaster(r) {
		RespondErrorCode(w, fmt.Errorf("master key required"), http.StatusUnauthorized)
		return
	}

	// null/absent field clears that override (nil = inherit env).
	var cfg models.GlobalMessageConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		RespondErrorCode(w, fmt.Errorf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if err := models.SetGlobalMessageConfig(cfg); err != nil {
		RespondErrorCode(w, fmt.Errorf("failed to save settings: %v", err), http.StatusInternalServerError)
		return
	}

	RespondSuccess(w, models.GetGlobalMessageConfig())
}

func registerCanonicalSettingsRoutes(r chi.Router) {
	r.Get("/settings", CanonicalSettingsGetController)
	r.Put("/settings", CanonicalSettingsPutController)
}
