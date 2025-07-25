package form

import (
	"html/template"
	"net/http"

	api "github.com/nocodeleaks/quepasa/api"
	models "github.com/nocodeleaks/quepasa/models"
)

type indexData struct {
	PageTitle string
}

// IndexHandler renders route GET "/"
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	_, err := GetFormUser(r)
	if err != nil {
		if err == models.ErrFormUnauthenticated {
			RedirectToLogin(w, r)
			return
		}

		api.RespondInterface(w, err)
		return
	}

	data := indexData{
		PageTitle: "Home",
	}

	templates := template.Must(template.ParseFiles("views/layouts/main.tmpl", "views/index.tmpl"))
	templates.ExecuteTemplate(w, "main", data)
}
