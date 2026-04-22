package viewmodel

import (
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"
)

// LoginPageData contains only the template fields required by the login screen.
// Keeping form page data in the form module avoids leaking HTML/view concerns
// into the shared domain package.
type LoginPageData struct {
	PageTitle string
	Version   string
}

// AccountPageData is the template model for the authenticated account screen.
// It intentionally depends on runtime/server details because this package sits
// at the UI edge, not in the core domain.
type AccountPageData struct {
	PageTitle                   string
	ErrorMessage                string
	Version                     string
	Servers                     map[string]*models.QpWhatsappServer
	User                        models.QpUser
	Options                     whatsapp.WhatsappOptionsExtended `json:"options,omitempty"`
	WMOptions                   whatsmeow.WhatsmeowOptions       `json:"wmoptions,omitempty"`
	HasSignalRActiveConnections bool
	HasMasterKey                bool
}

// SendPageData contains fields rendered by the manual send-message form.
type SendPageData struct {
	PageTitle    string
	MessageId    string
	ErrorMessage string
	Server       *models.QpServer
}

// ReceivePageData contains the receive/history form filters and rendered rows.
type ReceivePageData struct {
	PageTitle           string
	ErrorMessage        string
	Number              string
	Token               string
	DownloadPrefix      string
	FormAccountEndpoint string
	TimestampFilter     string
	LastFilter          string
	SearchFilter        string
	CategoryFilter      string
	TypeFilter          string
	ExceptionsFilter    string
	FromMeFilter        string
	FromHistoryFilter   string
	ChatIDFilter        string
	MessageIDFilter     string
	TrackIDFilter       string
	Messages            []whatsapp.WhatsappMessage
}

// Count returns the number of rendered messages for the receive template.
func (source ReceivePageData) Count() int {
	return len(source.Messages)
}

// VerifyPageData drives the verify/pairing screen rendered by the legacy form UI.
type VerifyPageData struct {
	PageTitle    string
	ErrorMessage string
	Bot          models.QPBot
	Protocol     string
	Host         string
	Destination  string
}

// WebHooksPageData contains the legacy form screen state for webhook management.
type WebHooksPageData struct {
	PageTitle    string
	ErrorMessage string
	Server       *models.QpWhatsappServer
	Webhooks     []*models.QpWebhook
}
