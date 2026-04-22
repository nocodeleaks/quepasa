package api

import (
	"net/http"

	environment "github.com/nocodeleaks/quepasa/environment"
	models "github.com/nocodeleaks/quepasa/models"
)

// LoginConfigController returns the public bootstrap payload used by SPA login screens.
//
// The current implementation intentionally keeps most branding fields empty because
// we have not imported PR #39 branding/environment customizations yet. The endpoint
// still exists now so a SPA client can discover basic runtime facts without scraping
// templates or depending on private settings.
func LoginConfigController(w http.ResponseWriter, r *http.Request) {
	appTitle := environment.Settings.General.AppTitle
	if appTitle == "" {
		appTitle = "QuePasa"
	}

	response := map[string]interface{}{
		// Stable fields already supported by the current environment model.
		"appTitle":     appTitle,
		"accountSetup": environment.Settings.General.AccountSetup,
		"version":      models.QpVersion,
		// Reserved compatibility fields expected by the SPA branch. They remain empty
		// until we decide to import branding/login customization support explicitly.
		"loginLogo":     "",
		"loginSubtitle": "",
		"loginWarning":  "",
		"loginFooter":   "",
		"loginLayout":   "center",
		"customCss":     "",
		"fontAwesome":   "",
		"googleFonts":   "",
		// SQLite migration discovery from PR #39 is not wired yet, so expose the
		// field with a conservative false value to keep the payload shape stable.
		"sqliteMigrationAvailable": false,
		"branding": map[string]interface{}{
			"title":          appTitle,
			"logo":           "",
			"favicon":        "",
			"primaryColor":   "",
			"secondaryColor": "",
			"accentColor":    "",
			"companyName":    "",
			"companyUrl":     "",
		},
	}

	RespondSuccess(w, response)
}
