package form

import (
	"net/http"

	api "github.com/nocodeleaks/quepasa/api"
	environment "github.com/nocodeleaks/quepasa/environment"
	models "github.com/nocodeleaks/quepasa/models"
)

// FormLoginJSONController returns login page customization for form clients.
func FormLoginJSONController(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"appTitle":      environment.Settings.General.AppTitle,
		"loginLogo":     environment.Settings.General.LoginLogo,
		"loginSubtitle": environment.Settings.General.LoginSubtitle,
		"loginWarning":  environment.Settings.General.LoginWarning,
		"loginFooter":   environment.Settings.General.LoginFooter,
		"loginLayout":   environment.Settings.General.LoginLayout,
		"customCss":     environment.Settings.General.LoginCustomCSS,
		"fontAwesome":   environment.Settings.General.LoginFontAwesome,
		"googleFonts":   environment.Settings.General.LoginGoogleFonts,
		"accountSetup":  environment.Settings.General.AccountSetup,
		"version":       models.QpVersion,
	}

	api.RespondSuccess(w, resp)
}
