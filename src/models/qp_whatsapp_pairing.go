package models

import (
	"context"

	"github.com/google/uuid"
	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

type QpWhatsappPairing struct {
	// Public token
	Token string `db:"token" json:"token" validate:"max=100"`

	// Whatsapp session id
	Wid string `db:"wid" json:"wid" validate:"max=255"`

	Username string `json:"username,omitempty"`

	HistorySyncDays uint32 `json:"historysyncdays,omitempty"`

	conn whatsapp.IWhatsappConnection `json:"-"`
}

func (source *QpWhatsappPairing) GetLogger() *log.Entry {
	if source.conn != nil && !source.conn.IsInterfaceNil() {
		return source.conn.GetLogger()
	}

	logentry := log.WithContext(context.Background())
	logentry = logentry.WithField(LogFields.Token, source.Token)
	logentry = logentry.WithField(LogFields.WId, source.Wid)
	return logentry
}

func (source *QpWhatsappPairing) OnPaired(wid string) {
	source.Wid = wid

	// if token was not setted
	// remember that user may want a different section for the same whatsapp
	if len(source.Token) == 0 {

		// updating token if from user
		if len(source.Username) > 0 {
			source.Token = source.GetUserToken()
		}
	}

	if source.conn != nil {
		source.conn.SetReconnect(true)
	}

	logentry := source.GetLogger()
	loglevel := logentry.Level
	logentry = logentry.WithField(LogFields.Token, source.Token)
	logentry = logentry.WithField(LogFields.WId, source.Wid)
	logentry.Level = loglevel

	if len(source.Username) == 0 {
		logentry.Error("whatsapp pairing, on paired, missing username")
		return
	}

	logentry.Info("whatsapp pairing, on paired, appending")
	server, err := WhatsappService.AppendPaired(source)
	if err != nil {
		logentry.Errorf("whatsapp pairing, on paired, append error: %s", err.Error())
		return
	}

	go server.EnsureReady()
}

func (source *QpWhatsappPairing) GetConnection() (whatsapp.IWhatsappConnection, error) {
	if source.conn == nil {
		conn, err := NewEmptyConnection(source.OnPaired)
		if err != nil {
			return nil, err
		}
		source.conn = conn
	}

	return source.conn, nil
}

// gets an existent token for same phone number and user, or create a new token
func (source *QpWhatsappPairing) GetUserToken() string {
	phone := library.GetPhoneByWId(source.Wid)

	logentry := source.GetLogger()
	logentry.Infof("whatsapp pairing, get user token, phone by wid: %s", phone)

	servers := WhatsappService.GetServersForUser(source.Username)
	for _, item := range servers {
		if item.GetNumber() == phone {
			return item.Token
		}
	}

	return uuid.New().String()
}
