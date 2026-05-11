package models

import (
	"fmt"
	"time"

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

func (source *QpWhatsappServer) DownloadData(id string) ([]byte, error) {
	msg, err := source.Handler.GetById(id)
	if err != nil {
		return nil, err
	}

	logentry := source.GetLogger()
	logentry = logentry.WithField(LogFields.MessageId, id)
	logentry.Info("downloading msg data")

	return source.connection.DownloadData(msg)
}

/*
<summary>

	Download attachment from msg id, optional use cached data or not

</summary>
*/
func (source *QpWhatsappServer) Download(id string, cache bool) (att *whatsapp.WhatsappAttachment, err error) {
	msg, err := source.Handler.GetById(id)
	if err != nil {
		return
	}

	logentry := source.GetLogger()
	logentry = logentry.WithField(LogFields.MessageId, id)
	logentry.Infof("downloading msg attachment, using cache: %v", cache)

	att, err = source.connection.Download(msg, cache)
	if err != nil {
		return
	}

	return
}

func (source *QpWhatsappServer) RevokeByPrefix(id string) (errors []error) {
	messages := source.Handler.GetByPrefix(id)
	for _, msg := range messages {
		if msg == nil {
			continue
		}
		if msg.Type == whatsapp.SystemMessageType {
			errors = append(errors, fmt.Errorf("system messages cannot be revoked"))
			continue
		}
		source.GetLogger().Infof("revoking msg by prefix %s", msg.Id)
		err := source.connection.Revoke(msg)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return
}

func (source *QpWhatsappServer) Revoke(id string) (err error) {
	msg, err := source.Handler.GetById(id)
	if err != nil {
		return
	}
	if msg != nil && msg.Type == whatsapp.SystemMessageType {
		return fmt.Errorf("system messages cannot be revoked")
	}

	source.GetLogger().Infof("revoking msg %s", id)
	return source.connection.Revoke(msg)
}

func (source *QpWhatsappServer) Edit(id string, newContent string) (err error) {
	msg, err := source.Handler.GetById(id)
	if err != nil {
		return
	}

	source.GetLogger().Infof("editing msg %s", id)
	return source.connection.Edit(msg, newContent)
}

func (source *QpWhatsappServer) MarkRead(id string) (err error) {
	msg, err := source.Handler.GetById(id)
	if err != nil {
		return
	}
	source.GetLogger().Infof("marking msg %s as read", id)
	return source.connection.MarkRead(msg)
}

func (server *QpWhatsappServer) GetMessages(timestamp time.Time) (messages []whatsapp.WhatsappMessage) {
	if !timestamp.IsZero() && timestamp.Unix() > 0 {
		err := server.connection.HistorySync(timestamp)
		if err != nil {
			logentry := server.GetLogger()
			logentry.Warnf("error on requested history sync: %s", err.Error())
		}
	}

	for _, item := range server.Handler.GetByTime(timestamp) {
		messages = append(messages, *item)
	}
	return
}

// Default send message method
func (source *QpWhatsappServer) SendMessage(msg *whatsapp.WhatsappMessage) (response whatsapp.IWhatsappSendResponse, err error) {
	logger := source.GetLogger()
	logger.Debugf("sending msg to: %s", msg.Chat.Id)

	conn, err := source.GetValidConnection()
	if err != nil {
		return
	}

	// Normalize Brazilian mobile phone number (handles 8/9-digit ambiguity).
	// Queries IsOnWhatsApp at most once per phone per session (cached in WhatsmeowContactMaps).
	if ENV.ShouldNormalizeBRPhone() {

		phone, _ := whatsapp.GetPhoneIfValid(msg.Chat.Id)
		if len(phone) > 0 {
			// Try remove-9: 9-digit (14 chars) → 8-digit variant (DDDs > 30)
			phoneWithout9, errRemove := library.RemoveDigit9IfElegible(phone)

			// Try add-9: 8-digit (13 chars) → 9-digit variant (all Brazilian DDDs)
			phoneWith9, errAdd := library.AddDigit9BRAllDDDs(phone)

			var variant string
			if errRemove == nil {
				variant = phoneWithout9
			} else if errAdd == nil {
				variant = phoneWith9
			}

			if len(variant) > 0 {
				contactManager := source.GetContactManager()
				valids, err := contactManager.IsOnWhatsApp(phone, variant)
				if err != nil {
					return nil, err
				}

				for _, valid := range valids {
					logger.Debugf("found valid destination: %s", valid)
					msg.Chat.Id = valid
					break
				}
			}
		}
	}

	// Trick to send audio with text, creating a new msg
	if msg.HasAttachment() {

		// Overriding filename with caption text if IMAGE or VIDEO
		if len(msg.Text) > 0 && msg.Type == whatsapp.AudioMessageType {

			// Copying and send text before file
			textMsg := *msg
			textMsg.Type = whatsapp.TextMessageType
			textMsg.Attachment = nil
			response, err = conn.Send(&textMsg)
			if err != nil {
				return
			} else {
				source.Handler.Message(&textMsg, "text and audio")
			}

			// updating id for audio message, if is set
			if len(msg.Id) > 0 {
				msg.Id = msg.Id + "-audio"
			}

			// removing message text, already sended ...
			msg.Text = ""
		}
	}

	// sending default msg
	response, err = conn.Send(msg)
	if err == nil {
		source.Handler.Message(msg, "server send")
	}
	return
}

func (source *QpWhatsappServer) GetProfilePicture(wid string, knowingId string) (picture *whatsapp.WhatsappProfilePicture, err error) {
	logger := source.GetLogger()
	logger.Debugf("getting info about profile picture for: %s, with id: %s", wid, knowingId)

	// future implement a rate control here, high volume of requests causing bans
	// studying rates ...

	contactManager := source.GetContactManager()
	return contactManager.GetProfilePicture(wid, knowingId)
}

// GetContacts retrieves contacts from WhatsApp
// Works with both active connection and stopped server (uses cached data automatically)
func (source *QpWhatsappServer) GetContacts() (contacts []whatsapp.WhatsappChat, err error) {
	contactManager := source.GetContactManager()
	contacts, err = contactManager.GetContacts()
	if err == nil {
		for index, contact := range contacts {
			contact.Id = library.TrimSessionIdFromWIdString(contact.Id)
			contacts[index] = contact
		}
	}

	return
}

func (source *QpWhatsappServer) IsOnWhatsApp(phones ...string) (registered []string, err error) {
	contactManager := source.GetContactManager()
	return contactManager.IsOnWhatsApp(phones...)
}

// GetGroupManager returns the group manager instance with lazy initialization
func (server *QpWhatsappServer) GetGroupManager() whatsapp.WhatsappGroupManagerInterface {
	if server.GroupManager == nil {
		server.GroupManager = NewQpGroupManager(server)
	}
	return server.GroupManager
}

// GetStatusManager returns the status manager instance with lazy initialization
func (server *QpWhatsappServer) GetStatusManager() whatsapp.WhatsappStatusManagerInterface {
	if server.StatusManager == nil {
		server.StatusManager = NewQpStatusManager(server)
	}
	return server.StatusManager
}

// GetContactManager returns the contact manager instance with lazy initialization
func (server *QpWhatsappServer) GetContactManager() whatsapp.WhatsappContactManagerInterface {
	if server.ContactManager == nil {
		server.ContactManager = NewQpContactManager(server)
	}
	return server.ContactManager
}

func (server *QpWhatsappServer) SendChatPresence(chatId string, presenceType whatsapp.WhatsappChatPresenceType) error {
	conn, err := server.GetValidConnection()
	if err != nil {
		return err
	}
	return conn.SendChatPresence(chatId, uint(presenceType))
}

func (server *QpWhatsappServer) GetLIDFromPhone(phone string) (string, error) {
	contactManager := server.GetContactManager()
	return contactManager.GetLIDFromPhone(phone)
}

func (server *QpWhatsappServer) GetPhoneFromLID(lid string) (string, error) {
	contactManager := server.GetContactManager()
	return contactManager.GetPhoneFromLID(lid)
}

// GetUserInfo retrieves user information for given JIDs
func (server *QpWhatsappServer) GetUserInfo(jids []string) ([]interface{}, error) {
	contactManager := server.GetContactManager()
	return contactManager.GetUserInfo(jids)
}
