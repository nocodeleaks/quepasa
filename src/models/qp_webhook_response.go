package models

// Requisição no formato QuePasa
// Utilizada na API do QuePasa para atualizar um WebHook de algum BOT
type QpWebhookResponse struct {
	QpResponse
	Affected uint         `json:"affected,omitempty"` // items affected
	Webhooks []*QpWebhook `json:"webhooks,omitempty"` // current items
}
