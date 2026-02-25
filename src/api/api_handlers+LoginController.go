package api

import (
	"net/http"

	environment "github.com/nocodeleaks/quepasa/environment"
	models "github.com/nocodeleaks/quepasa/models"
	log "github.com/sirupsen/logrus"
)

// LoginConfigController returns login page customizations from environment
func LoginConfigController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Infof("LoginConfigController called from %s", r.RemoteAddr)
	response := map[string]interface{}{
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
		"sqliteMigrationAvailable": func() bool {
			params := environment.Settings.Database.GetDBParameters()
			if params.Driver == "sqlite3" {
				return false
			}
			_, ok := models.CheckLocalSqliteExists()
			return ok
		}(),
		"branding": map[string]interface{}{
			"title":          environment.Settings.Branding.Title,
			"logo":           environment.Settings.Branding.Logo,
			"favicon":        environment.Settings.Branding.Favicon,
			"primaryColor":   environment.Settings.Branding.PrimaryColor,
			"secondaryColor": environment.Settings.Branding.SecondaryColor,
			"accentColor":    environment.Settings.Branding.AccentColor,
			"companyName":    environment.Settings.Branding.CompanyName,
			"companyUrl":     environment.Settings.Branding.CompanyUrl,
		},
	}
	RespondSuccess(w, response)
}
