package models

import (
	"fmt"
	"strings"
)

// Webhook model
type QpDataWebhooks struct {
	Webhooks []*QpWebhook `json:"webhooks,omitempty"`
	context  string
	db       QpDataWebhooksInterface
}

// Fill start memory cache
func (source *QpDataWebhooks) WebhookFill(info *QpServer, db QpDataWebhooksInterface) (err error) {
	webhooks := []*QpWebhook{}
	source.context = info.Token
	source.db = db

	logentry := info.GetLogger()
	logentry.Trace("getting webhooks from database")

	whooks, err := source.db.FindAll(source.context)
	if err != nil {
		logentry.Errorf("cannot find webhooks: %s", err.Error())
		return
	}

	for _, element := range whooks {

		// logging for running webhooks
		webHookLogEntry := logentry.WithField(LogFields.Url, element.Url)
		element.LogEntry = webHookLogEntry

		element.Wid = info.Wid
		webhooks = append(webhooks, element.QpWebhook)
	}

	logentry.Debugf("%v webhook(s) found", len(webhooks))

	// updating
	source.Webhooks = webhooks
	return
}

func (source *QpDataWebhooks) WebhookAddOrUpdate(webhook *QpWebhook) (affected uint, err error) {

	if webhook == nil || len(webhook.Url) == 0 {
		err = fmt.Errorf("empty or nil webhook")
		return
	}

	botWHook, err := source.db.Find(source.context, webhook.Url)
	if err != nil {
		return
	}

	if botWHook != nil {
		botWHook.ForwardInternal = webhook.ForwardInternal
		botWHook.TrackId = webhook.TrackId
		botWHook.Groups = webhook.Groups
		botWHook.ReadReceipts = webhook.ReadReceipts
		botWHook.Broadcasts = webhook.Broadcasts
		botWHook.Extra = webhook.Extra

		err = source.db.Update(botWHook)
		if err != nil {
			return
		}
	} else {
		dbWebhook := &QpServerWebhook{
			Context:   source.context,
			QpWebhook: webhook,
		}
		err = source.db.Add(dbWebhook)
		if err != nil {
			return
		}
	}

	exists := false
	for index, element := range source.Webhooks {
		if element.Url == webhook.Url {
			source.Webhooks = append(source.Webhooks[:index], source.Webhooks[index+1:]...) // remove
			source.Webhooks = append(source.Webhooks, webhook)                              // append a clean one
			exists = true
			affected++
		}
	}

	if !exists {
		source.Webhooks = append(source.Webhooks, webhook)
		affected++
	}

	return
}

func (source *QpDataWebhooks) WebhookRemove(url string) (affected uint, err error) {
	i := 0 // output index
	for _, element := range source.Webhooks {
		if len(url) == 0 || strings.Contains(element.Url, url) {
			err = source.db.Remove(source.context, element.Url)
			if err == nil {
				affected++
			} else {
				source.Webhooks[i] = element
				i++
				break
			}
		} else {
			source.Webhooks[i] = element
			i++
		}
	}

	for j := i; j < len(source.Webhooks); j++ {
		source.Webhooks[j] = nil
	}
	source.Webhooks = source.Webhooks[:i]
	return
}

func (source *QpDataWebhooks) WebhookClear() (err error) {
	return source.db.Clear(source.context)
}

/*
func (source *QpDataWebhooks) WebhookFailure(url string) {
	log.Infof("failure on webhook from: %s", url)
	for index, element := range source.Webhooks {
		if element.Url == url {
			//source.Webhooks[index].Failure = &time.Time{}
		}
	}
}
*/
