package models

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type QpDataServerWebhookSql struct {
	db *sqlx.DB
}

func (source QpDataServerWebhookSql) Find(context string, url string) (response *QpServerWebhook, err error) {
	var result []QpServerWebhook
	err = source.db.Select(&result, "SELECT * FROM webhooks WHERE context = ? AND url = ?", context, url)
	if err != nil {
		return
	}

	// adjust extra information
	for _, element := range result {
		element.ParseExtra()
		response = &element
		break
	}

	return
}

func (source QpDataServerWebhookSql) FindAll(context string) ([]*QpServerWebhook, error) {
	result := []*QpServerWebhook{}
	err := source.db.Select(&result, "SELECT * FROM webhooks WHERE context = ?", context)

	// adjust extra information
	for _, element := range result {
		element.ParseExtra()
	}
	return result, err
}

func (source QpDataServerWebhookSql) All() ([]*QpServerWebhook, error) {
	result := []*QpServerWebhook{}
	err := source.db.Select(&result, "SELECT * FROM webhooks")

	// adjust extra information
	for _, element := range result {
		element.ParseExtra()
	}
	return result, err
}

func (source QpDataServerWebhookSql) Add(element *QpServerWebhook) error {
	query := `INSERT OR IGNORE INTO webhooks (context, url, forwardinternal, trackid, readreceipts, groups, broadcasts, extra) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := source.db.Exec(query, element.Context, element.Url, element.ForwardInternal, element.TrackId, element.ReadReceipts, element.Groups, element.Broadcasts, element.GetExtraText())
	return err
}

func (source QpDataServerWebhookSql) Update(element *QpServerWebhook) error {
	query := `UPDATE webhooks SET forwardinternal = ?, trackid = ?, readreceipts = ?, groups = ?, broadcasts = ?, extra = ? WHERE context = ? AND url = ?`
	_, err := source.db.Exec(query, element.ForwardInternal, element.TrackId, element.ReadReceipts, element.Groups, element.Broadcasts, element.GetExtraText(), element.Context, element.Url)
	return err
}

func (source QpDataServerWebhookSql) UpdateContext(element *QpServerWebhook, context string) error {
	query := `UPDATE webhooks SET context = ? WHERE context = ? AND url = ?`
	_, err := source.db.Exec(query, context, element.Context, element.Url)
	if err != nil {
		element.Context = context
	}
	return err
}

func (source QpDataServerWebhookSql) Remove(context string, url string) error {
	query := `DELETE FROM webhooks WHERE context = ? AND url = ?`
	_, err := source.db.Exec(query, context, url)
	return err
}

func (source QpDataServerWebhookSql) Clear(context string) error {
	query := `DELETE FROM webhooks WHERE context = ?`
	_, err := source.db.Exec(query, context)
	return err
}
