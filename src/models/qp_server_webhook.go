package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Webhook model
type QpServerWebhook struct {
	Context string                  `db:"context" json:"context"`
	db      QpDataWebhooksInterface `json:"-"`

	*QpWebhook
}

func (source *QpServerWebhook) Find(context string, url string) (*QpServerWebhook, error) {
	return source.db.Find(context, url)
}

func (source *QpServerWebhook) FindAll(context string) ([]*QpServerWebhook, error) {
	return source.db.FindAll(context)
}

func (source *QpServerWebhook) All() ([]*QpServerWebhook, error) {
	return source.db.All()
}

// passing extra info as json valid or default string
func (source *QpServerWebhook) GetExtraText() string {
	extraJson, err := json.Marshal(&source.Extra)
	if err != nil {
		return fmt.Sprintf("%v", source.Extra)
	} else {
		return string(extraJson)
	}
}

// trying to get interface from json string or default string
func (source *QpServerWebhook) ParseExtra() {
	extraText := fmt.Sprintf("%v", source.Extra)

	var extraJson interface{}
	err := json.Unmarshal([]byte(extraText), &extraJson)
	if err != nil {
		source.Extra = extraText
	} else {
		source.Extra = extraJson
	}
}

func (source *QpServerWebhook) Add(element *QpServerWebhook) error {
	return source.db.Add(element)
}

func (source *QpServerWebhook) Remove(context string, url string) error {
	return source.db.Remove(context, url)
}

func (source *QpServerWebhook) Clear(context string) error {
	return source.db.Clear(context)
}

// Implement QpWebhookInterface

func (source *QpServerWebhook) GetUrl() string {
	return source.Url
}

func (source *QpServerWebhook) GetFailure() *time.Time {
	return source.Failure
}
