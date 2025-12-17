package form

import (
	"fmt"
	"html/template"
	"net/http"

	api "github.com/nocodeleaks/quepasa/api"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

var funcMap = template.FuncMap{"safeURL": SafeURL}

func SafeURL(url string) template.URL {
	return template.URL(url)
}

// FormReceiveController renders route GET "/form/server/{token}/receive"
func FormReceiveController(w http.ResponseWriter, r *http.Request) {
	data := models.QPFormReceiveData{PageTitle: "Receive - Quepasa", FormAccountEndpoint: FormAccountEndpoint}

	server, err := GetServerFromRequest(r)
	if err != nil {
		data.ErrorMessage = err.Error()
	}

	if server != nil {
		data.Number = server.GetWId()
		data.Token = server.Token
		data.DownloadPrefix = GetDownloadPrefix(server.Token)

		// Evitando tentativa de download de anexos sem o bot estar devidamente sincronizado
		status := server.GetStatus()
		if status != whatsapp.Ready {
			data.ErrorMessage = fmt.Sprintf("server (%s) not ready yet ! current status: %s; %s", server.Wid, status.String(), data.ErrorMessage)
		}

		timestamp, err := api.GetTimestamp(r)
		if err != nil {
			data.ErrorMessage = fmt.Sprintf("%s; %s", err.Error(), data.ErrorMessage)
		}

		messages := api.GetOrderedMessages(server, timestamp)
		data.Messages = messages
	}

	templates := template.New("receive")
	templates = templates.Funcs(funcMap)
	templates, err = templates.ParseFiles(
		GetViewPath("layouts/main.tmpl"),
		GetViewPath("bot/receive.tmpl"))
	if err != nil {
		data.ErrorMessage = err.Error()
	}
	templates.ExecuteTemplate(w, "main", data)
}
