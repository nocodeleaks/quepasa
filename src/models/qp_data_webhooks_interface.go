package models

type QpDataWebhooksInterface interface {
	Find(context string, url string) (*QpServerWebhook, error)
	FindAll(context string) ([]*QpServerWebhook, error)
	All() ([]*QpServerWebhook, error)
	Add(element *QpServerWebhook) error
	Update(element *QpServerWebhook) error
	UpdateContext(element *QpServerWebhook, context string) error
	Remove(context string, url string) error
	Clear(context string) error
}
