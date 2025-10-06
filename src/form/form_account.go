package form

import (
	"html/template"
	"net/http"

	"github.com/nbutton23/zxcvbn-go"
	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
)

func renderSetupForm(w http.ResponseWriter, data FormSetupData) {
	templates := template.Must(template.ParseFiles(GetViewPath("layouts/main.tmpl"), GetViewPath("setup.tmpl")))
	templates.ExecuteTemplate(w, "main", data)
}

// SetupFormHandler renders route GET "/setup"
func SetupFormHandler(w http.ResponseWriter, r *http.Request) {
	/* temporarily removed to permit multiple users
	count, err := WhatsappService.DB.User.Count()
	if count > 0 || err != nil {
		RedirectToLogin(w, r)
		return
	}
	*/
	data := FormSetupData{
		PageTitle: "Setup",
	}

	renderSetupForm(w, data)
}

// SetupHandler renders route POST "/setup"
func SetupHandler(w http.ResponseWriter, r *http.Request) {
	/* temporarily removed to permit multiple users
	count, err := WhatsappService.DB.User.Count()
	if count > 0 || err != nil {
		RedirectToLogin(w, r)
		return
	}
	*/
	data := FormSetupData{
		PageTitle: "Setup",
	}

	r.ParseForm()
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	passwordConfirm := r.Form.Get("passwordConfirm")

	if email == "" || password == "" {
		data.ErrorMessage = "Email and password are required"
		data.EmailError = true
		renderSetupForm(w, data)
		return
	}

	data.Email = email

	if !library.IsValidEMail(email) {
		data.ErrorMessage = "Email is invalid"
		data.EmailInvalidError = true
		renderSetupForm(w, data)
		return
	}

	if password != passwordConfirm {
		data.ErrorMessage = "Passwords don't match"
		data.PasswordMatchError = true
		renderSetupForm(w, data)
		return
	}

	res := zxcvbn.PasswordStrength(password, nil)
	if res.Score < 1 {
		data.ErrorMessage = "Password is too weak"
		data.PasswordStrengthError = true
		data.PasswordCrackTime = res.CrackTimeDisplay
		renderSetupForm(w, data)
		return
	}

	exists, err := models.WhatsappService.DB.Users.Exists(email)
	if err != nil {
		data.ErrorMessage = err.Error()
		renderSetupForm(w, data)
		return
	}

	if exists {
		data.UserExistsError = true
		renderSetupForm(w, data)
		return
	}

	_, err = models.WhatsappService.DB.Users.Create(email, password)
	if err != nil {
		data.ErrorMessage = err.Error()
		renderSetupForm(w, data)
		return
	}

	RedirectToLogin(w, r)
}
