package models

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

/*
<summary>

	Save changes on database

</summary>
*/
func (source *QpWhatsappServer) Save(reason string) (err error) {
	logger := source.GetLogger()

	logger.Infof("saving server info, reason: %s, json: %+v", reason, source)
	ok, err := source.db.Exists(source.Token)
	if err != nil {
		log.Errorf("error on checking existent server: %s", err.Error())
		return
	}

	// updating timestamp
	currentTime := time.Now().UTC()
	source.Timestamp = currentTime
	source.Timestamps.Update = currentTime

	if ok {
		logger.Debugf("updating server info: %+v", source)
		err = source.db.Update(source.QpServer)
		if err != nil {
			logger.Errorf("failed to update server in database (token=%s wid=%s): %v", source.Token, source.GetWId(), err)
		}
		return err
	} else {
		logger.Debugf("adding server info: %+v", source)
		err = source.db.Add(source.QpServer)
		if err != nil {
			logger.Errorf("failed to insert server in database (token=%s wid=%s): %v", source.Token, source.GetWId(), err)
		}
		return err
	}
}

func (server *QpWhatsappServer) MarkVerified(value bool) (err error) {
	if server.Verified != value {
		server.Verified = value

		reason := fmt.Sprintf("mark verified as %v", value)
		return server.Save(reason)
	}
	return nil
}

func (server *QpWhatsappServer) ClearConnectionIssue(reason string) error {
	if server == nil {
		return nil
	}

	server.RemoveMetadataValue(connectionDiagnosticMetadataKey)

	if reason == "" {
		reason = "clear connection issue"
	}

	return server.Save(reason)
}

func (server *QpWhatsappServer) RecordDisconnect(cause, details string) error {
	if server == nil {
		return nil
	}

	now := time.Now().UTC()
	server.SetMetadataValue(connectionDiagnosticMetadataKey, &QpConnectionDiagnostic{
		Code:              deriveDisconnectIssueCode(cause),
		Message:           buildDisconnectIssueMessage(cause, details),
		OccurredAt:        &now,
		RequiresReauth:    false,
		DisconnectCause:   cause,
		DisconnectDetails: details,
	})

	return server.Save(fmt.Sprintf("record disconnect: %s", cause))
}

func (server *QpWhatsappServer) RecordLogout(reason string) error {
	if server == nil {
		return nil
	}

	now := time.Now().UTC()
	server.Verified = false
	server.SetMetadataValue(connectionDiagnosticMetadataKey, &QpConnectionDiagnostic{
		Code:              deriveLogoutIssueCode(reason),
		Message:           buildLogoutIssueMessage(reason),
		OccurredAt:        &now,
		RequiresReauth:    true,
		DisconnectCause:   "logged_out",
		DisconnectDetails: reason,
		LogoutReason:      reason,
	})

	return server.Save(fmt.Sprintf("record logout: %s", reason))
}

func (source *QpWhatsappServer) ToggleDevel() (handle bool, err error) {
	source.Devel = !source.Devel

	logentry := source.GetLogger()
	if source.Devel {
		logentry.Level = log.DebugLevel
	} else {
		logentry.Level = log.InfoLevel
	}

	reason := fmt.Sprintf("toggle devel: %v", source.Devel)
	return source.Devel, source.Save(reason)
}

// delete this whatsapp server and underlaying connection
func (server *QpWhatsappServer) Delete(cause string) (err error) {
	if server == nil {
		return fmt.Errorf("whatsapp server, delete error: nil server")
	}

	if server.db == nil {
		return fmt.Errorf("whatsapp server, delete error: server database not configured")
	}

	previousState := server.GetState()
	previousIntent := server.Intent
	dispatchingSnapshot := cloneDispatchings(server.QpDataDispatching.Dispatching)

	server.Intent = SessionIntentDelete

	defer func() {
		if err != nil {
			server.Intent = previousIntent
		}
	}()

	if server.connection != nil {
		err = server.connection.Delete()
		if err != nil {
			return fmt.Errorf("whatsapp server, delete connection, error: %s", err.Error())
		}

		server.connection = nil
	}

	err = server.clearDispatchingsForDelete()
	if err != nil {
		return fmt.Errorf("whatsapp server, dispatching clear, error: %s", err.Error())
	}

	err = server.db.Delete(server.Token)
	if err != nil {
		restoreErr := server.restoreDispatchingsAfterFailedDelete(dispatchingSnapshot)
		if restoreErr != nil {
			return fmt.Errorf("whatsapp server, database delete connection, error: %s; dispatching restore error: %s", err.Error(), restoreErr.Error())
		}
		return fmt.Errorf("whatsapp server, database delete connection, error: %s", err.Error())
	}

	if len(dispatchingSnapshot) > 0 {
		deleteEvent := NewServerDeletedEvent(server, cause, &previousState)
		dispatchErr := PostToDispatchings(server, dispatchingSnapshot, deleteEvent)
		if dispatchErr != nil {
			server.GetLogger().Errorf("error dispatching delete event: %s", dispatchErr.Error())
		}
	}

	return nil
}

func cloneDispatchings(dispatchings []*QpDispatching) []*QpDispatching {
	if len(dispatchings) == 0 {
		return nil
	}

	cloned := make([]*QpDispatching, len(dispatchings))
	copy(cloned, dispatchings)
	return cloned
}

func (server *QpWhatsappServer) clearDispatchingsForDelete() error {
	if server == nil {
		return nil
	}

	if server.QpDataDispatching.db == nil {
		server.QpDataDispatching.Dispatching = server.QpDataDispatching.Dispatching[:0]
		return nil
	}

	return server.DispatchingClear()
}

func (server *QpWhatsappServer) restoreDispatchingsAfterFailedDelete(snapshot []*QpDispatching) error {
	if server == nil {
		return nil
	}

	if len(snapshot) == 0 {
		server.QpDataDispatching.Dispatching = nil
		return nil
	}

	if server.QpDataDispatching.db == nil {
		server.QpDataDispatching.Dispatching = cloneDispatchings(snapshot)
		return nil
	}

	server.QpDataDispatching.Dispatching = server.QpDataDispatching.Dispatching[:0]
	for _, dispatching := range snapshot {
		if _, err := server.DispatchingAddOrUpdate(dispatching); err != nil {
			return err
		}
	}

	return nil
}
