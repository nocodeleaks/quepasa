package controllers

import (
	"fmt"
	"html/template"
	"net/http"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// FormReceiveController renders route GET "/form/server/{token}/receive"
func FormReceiveController(w http.ResponseWriter, r *http.Request) {
	data := models.QPFormReceiveData{PageTitle: "Receive", FormAccountEndpoint: FormAccountEndpoint}

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

		timestamp, err := GetTimestamp(r)
		if err != nil {
			data.ErrorMessage = fmt.Sprintf("%s; %s", err.Error(), data.ErrorMessage)
		}

		messages := GetOrderedMessages(server, timestamp)
		data.Messages = messages
	}

	templates := template.Must(template.ParseFiles(
		"views/layouts/main.tmpl",
		"views/bot/receive.tmpl"))
	templates.ExecuteTemplate(w, "main", data)
}
