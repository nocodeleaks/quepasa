package form

import (
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	api "github.com/nocodeleaks/quepasa/api"
	metrics "github.com/nocodeleaks/quepasa/metrics"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

func FormSendController(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		controllerHttpGet(w, r)
	case "POST":
		controllerHttpPost(w, r)
	}
}

// Renders route GET "/bot/{token}/send"
func controllerHttpGet(w http.ResponseWriter, r *http.Request) {
	data := models.QPFormSendData{PageTitle: "Send"}

	server, err := GetServerFromRequest(r)
	if err != nil {
		data.ErrorMessage = err.Error()
		renderSendForm(w, data)
		return
	} else {
		data.Server = server.QpServer
	}

	renderSendForm(w, data)
}

// Renders route POST "/bot/{token}/send"
// Vindo do formulário de testes
func controllerHttpPost(w http.ResponseWriter, r *http.Request) {
	data := models.QPFormSendData{PageTitle: "Send"}

	server, err := GetServerFromRequest(r)
	if err != nil {
		data.ErrorMessage = err.Error()
		renderSendForm(w, data)
		return
	} else {
		data.Server = server.QpServer
	}

	logentry := server.GetLogger()
	attachment, err := GetAttachFromUploadedFile(r, logentry)
	if err != nil {
		data.ErrorMessage = err.Error()
		renderSendForm(w, data)
		return
	}

	r.ParseForm()
	recipient := r.Form.Get("recipient")
	message := r.Form.Get("message")

	msg, err := models.ToWhatsappMessage(recipient, message, attachment)
	if err != nil {
		api.RespondServerError(server, w, err)
		return
	}

	msg.Id = r.Form.Get("id")
	_, err = server.SendMessage(msg)
	if err != nil {
		api.RespondServerError(server, w, err)
		return
	}

	data.MessageId = msg.GetId()

	// Increment counter statistics
	metrics.MessagesSent.Inc()

	renderSendForm(w, data)
}

func GetAttachFromUploadedFile(r *http.Request, logentry *log.Entry) (attach *whatsapp.WhatsappAttachment, err error) {
	logentry.Trace("form post, checking for file")

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		return
	}

	// FormFile returns the first file for the given key `attachment`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, reader, err := r.FormFile("attachment")
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			err = nil
			return
		}
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return
	}

	attach = &whatsapp.WhatsappAttachment{}
	attach.SetContent(&content)
	attach.Mimetype = reader.Header.Get("Content-Type")

	InformedLength := reader.Header.Get("Content-Length")
	if len(InformedLength) > 0 {
		length, err := strconv.ParseUint(InformedLength, 10, 64)
		if err == nil {
			attach.FileLength = length
		}
	} else {
		attach.FileLength = uint64(reader.Size)
	}

	attach.FileName = reader.Filename

	result := &models.QpToWhatsappAttachment{Attach: attach}
	result.AttachSecureAndCustomize()
	result.AttachImageTreatment()
	result.AttachAudioTreatment()
	for _, debug := range result.Debug {
		logentry.Debug(debug)
	}

	return
}

func renderSendForm(w http.ResponseWriter, data models.QPFormSendData) {
	templates := template.Must(template.ParseFiles("views/layouts/main.tmpl", "views/bot/send.tmpl"))
	templates.ExecuteTemplate(w, "main", data)
}
