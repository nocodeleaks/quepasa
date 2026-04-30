package runtime

import (
	"errors"
	"fmt"
	"strings"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type SessionConfigurationPatch struct {
	Groups       *whatsapp.WhatsappBoolean
	Broadcasts   *whatsapp.WhatsappBoolean
	ReadReceipts *whatsapp.WhatsappBoolean
	Calls        *whatsapp.WhatsappBoolean
	ReadUpdate   *whatsapp.WhatsappBoolean
	Devel        *bool
}

var ErrNilSession = errors.New("session is nil")
var ErrNilSessionInfo = errors.New("session info is nil")
var ErrSessionServiceUnavailable = errors.New("whatsapp service is not initialised")

// StartSession is the explicit runtime entry point for session startup.
func StartSession(session *models.QpWhatsappSession) error {
	if session == nil {
		return ErrNilSession
	}

	return session.Start()
}

// StopSession is the explicit runtime entry point for session shutdown.
func StopSession(session *models.QpWhatsappSession, cause string) error {
	if session == nil {
		return ErrNilSession
	}

	return session.Stop(cause)
}

// RestartSession is the explicit runtime entry point for session restart.
func RestartSession(session *models.QpWhatsappSession) error {
	if session == nil {
		return ErrNilSession
	}

	return session.Restart()
}

// RestartSessionAsync mirrors the fire-and-forget restart flow used by HTTP error handling.
func RestartSessionAsync(session *models.QpWhatsappSession) {
	if session == nil {
		return
	}

	go func() {
		_ = RestartSession(session)
	}()
}

// SendSessionMessage is the explicit runtime entry point for outbound message delivery.
func SendSessionMessage(session *models.QpWhatsappSession, msg *whatsapp.WhatsappMessage) (whatsapp.IWhatsappSendResponse, error) {
	if session == nil {
		return nil, ErrNilSession
	}

	return session.SendMessage(msg)
}

// SaveSession persists the current session state with an explicit reason.
func SaveSession(session *models.QpWhatsappSession, reason string) error {
	if session == nil {
		return ErrNilSession
	}

	return session.Save(reason)
}

// CreateSessionRecord creates and persists a session record without opening a live connection.
func CreateSessionRecord(info *models.QpServer, reason string) (*models.QpWhatsappSession, error) {
	if info == nil {
		return nil, ErrNilSessionInfo
	}

	session, err := models.WhatsappService.AppendNewSession(info)
	if err != nil {
		return nil, err
	}

	if err := SaveSession(session, reason); err != nil {
		if models.WhatsappService != nil && models.WhatsappService.Servers != nil {
			delete(models.WhatsappService.Servers, info.Token)
		}
		return nil, err
	}

	return session, nil
}

// BuildSessionRecord assembles a new persisted-session record before it is loaded
// into the runtime service.
func BuildSessionRecord(token string, username string, patch *SessionConfigurationPatch) *models.QpServer {
	info := &models.QpServer{Token: token}
	info.SetUser(username)

	if patch == nil {
		return info
	}

	if patch.Groups != nil {
		info.Groups = *patch.Groups
	}
	if patch.Broadcasts != nil {
		info.Broadcasts = *patch.Broadcasts
	}
	if patch.ReadReceipts != nil {
		info.ReadReceipts = *patch.ReadReceipts
	}
	if patch.Calls != nil {
		info.Calls = *patch.Calls
	}
	if patch.ReadUpdate != nil {
		info.ReadUpdate = *patch.ReadUpdate
	}
	if patch.Devel != nil {
		info.Devel = *patch.Devel
	}

	return info
}

// FindLiveSessionByToken looks up a live in-memory session by token without
// materializing persisted records.
func FindLiveSessionByToken(token string) (*models.QpWhatsappSession, bool) {
	if models.WhatsappService == nil {
		return nil, false
	}

	session, err := models.WhatsappService.FindSessionByToken(token)
	if err != nil {
		return nil, false
	}

	return session, true
}

// GetLiveSessionByToken returns a live in-memory session by token using the
// standard not-found error contract.
func GetLiveSessionByToken(token string) (*models.QpWhatsappSession, error) {
	session, ok := FindLiveSessionByToken(token)
	if !ok {
		return nil, models.ErrSessionNotFound
	}

	return session, nil
}

// GetFirstReadySession returns the first live session currently ready to serve requests.
func GetFirstReadySession() (*models.QpWhatsappSession, error) {
	if models.WhatsappService == nil {
		return nil, models.ErrSessionNotFound
	}

	for _, session := range models.WhatsappService.Servers {
		if session != nil && session.GetStatus() == whatsapp.Ready {
			return session, nil
		}
	}

	return nil, models.ErrSessionNotFound
}

// ListLiveSessions returns all currently cached in-memory sessions.
func ListLiveSessions() []*models.QpWhatsappSession {
	if models.WhatsappService == nil {
		return nil
	}

	sessions := make([]*models.QpWhatsappSession, 0, len(models.WhatsappService.Servers))
	for _, session := range models.WhatsappService.Servers {
		if session != nil {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// ListLiveSessionsForUser returns all cached in-memory sessions owned by a given user.
func ListLiveSessionsForUser(username string) []*models.QpWhatsappSession {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil
	}

	allSessions := ListLiveSessions()
	sessions := make([]*models.QpWhatsappSession, 0, len(allSessions))
	for _, session := range allSessions {
		if session != nil && session.GetUser() == username {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// GetSessionDownloadPrefix returns the stable internal download prefix for a live session.
func GetSessionDownloadPrefix(token string) (string, error) {
	session, err := GetLiveSessionByToken(token)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("/bot/%s/download", session.Token), nil
}

// ListPersistedSessionRecords returns all persisted session records from the database.
func ListPersistedSessionRecords() ([]*models.QpServer, error) {
	if models.WhatsappService == nil || models.WhatsappService.DB == nil || models.WhatsappService.DB.Servers == nil {
		return nil, fmt.Errorf("database service not initialized")
	}

	return models.WhatsappService.DB.Servers.FindAll(), nil
}

// FindPersistedSessionRecord resolves one persisted session record by token, with
// a case-insensitive fallback scan to preserve current API behavior.
func FindPersistedSessionRecord(token string) (*models.QpServer, error) {
	resolvedToken := strings.TrimSpace(token)
	if resolvedToken == "" {
		return nil, fmt.Errorf("missing token parameter")
	}

	records, err := ListPersistedSessionRecords()
	if err != nil {
		return nil, err
	}

	record, err := models.WhatsappService.DB.Servers.FindByToken(resolvedToken)
	if err == nil {
		return record, nil
	}

	for _, candidate := range records {
		if candidate == nil {
			continue
		}

		if strings.EqualFold(candidate.Token, resolvedToken) {
			return candidate, nil
		}
	}

	return nil, err
}

// FindPersistedUser resolves one persisted user by username.
func FindPersistedUser(username string) (*models.QpUser, error) {
	if models.WhatsappService == nil || models.WhatsappService.DB == nil || models.WhatsappService.DB.Users == nil {
		return nil, fmt.Errorf("user service not initialized")
	}

	return models.WhatsappService.DB.Users.Find(strings.TrimSpace(username))
}

// AuthenticateUser resolves and authenticates a persisted user.
func AuthenticateUser(username string, password string) (*models.QpUser, error) {
	if models.WhatsappService == nil {
		return nil, fmt.Errorf("user service not initialized")
	}

	return models.WhatsappService.GetUser(strings.TrimSpace(username), password)
}

// UpdatePersistedUserPassword updates the stored password after confirming the user exists.
func UpdatePersistedUserPassword(username string, password string) error {
	if models.WhatsappService == nil || models.WhatsappService.DB == nil || models.WhatsappService.DB.Users == nil {
		return fmt.Errorf("user service not initialized")
	}

	username = strings.TrimSpace(username)
	exists, err := models.WhatsappService.DB.Users.Exists(username)
	if err != nil {
		return fmt.Errorf("error on database check if user exists: %s", err.Error())
	}

	if !exists {
		return fmt.Errorf("user not found: %s", username)
	}

	if err := models.WhatsappService.DB.Users.UpdatePassword(username, password); err != nil {
		return fmt.Errorf("error on database updating password: %s", err.Error())
	}

	return nil
}

// CountPersistedUsers returns the total number of persisted users.
func CountPersistedUsers() (int, error) {
	if models.WhatsappService == nil || models.WhatsappService.DB == nil || models.WhatsappService.DB.Users == nil {
		return 0, fmt.Errorf("user service not initialized")
	}

	return models.WhatsappService.DB.Users.Count()
}

// ListPersistedUsers returns all persisted users from the configured store.
func ListPersistedUsers() ([]*models.QpUser, error) {
	if models.WhatsappService == nil || models.WhatsappService.DB == nil || models.WhatsappService.DB.Users == nil {
		return nil, fmt.Errorf("user service not initialized")
	}

	return models.WhatsappService.DB.Users.FindAll()
}

// CreatePersistedUser creates a new persisted user after trimming the username.
func CreatePersistedUser(username string, password string) (*models.QpUser, error) {
	if models.WhatsappService == nil || models.WhatsappService.DB == nil || models.WhatsappService.DB.Users == nil {
		return nil, fmt.Errorf("user service not initialized")
	}

	return models.WhatsappService.DB.Users.Create(strings.TrimSpace(username), password)
}

// DeletePersistedUser removes a persisted user by username.
func DeletePersistedUser(username string) error {
	if models.WhatsappService == nil || models.WhatsappService.DB == nil || models.WhatsappService.DB.Users == nil {
		return fmt.Errorf("user service not initialized")
	}

	return models.WhatsappService.DB.Users.Delete(strings.TrimSpace(username))
}

// FindPersistedDispatching returns all persisted dispatching entries for a session token.
func FindPersistedDispatching(token string) ([]*models.QpServerDispatching, error) {
	if models.WhatsappService == nil || models.WhatsappService.DB == nil || models.WhatsappService.DB.Dispatching == nil {
		return nil, fmt.Errorf("dispatching service not initialized")
	}

	return models.WhatsappService.DB.Dispatching.FindAll(token)
}

// GetConversationLabelStore resolves the configured conversation-label store.
func GetConversationLabelStore() (models.QpDataConversationLabelsInterface, error) {
	if models.WhatsappService == nil || models.WhatsappService.DB == nil || models.WhatsappService.DB.ConversationLabels == nil {
		return nil, fmt.Errorf("conversation labels service not initialized")
	}

	return models.WhatsappService.DB.ConversationLabels, nil
}

// DiagnoseOrphanedSessions exposes orphaned-session diagnostics through the
// runtime layer.
func DiagnoseOrphanedSessions() (*models.RestoreReport, error) {
	if models.WhatsappService == nil {
		return nil, ErrSessionServiceUnavailable
	}

	return models.WhatsappService.DiagnoseOrphaned()
}

// RestoreOrphanedSessions exposes automatic orphan restore through the runtime layer.
func RestoreOrphanedSessions() (*models.RestoreReport, error) {
	if models.WhatsappService == nil {
		return nil, ErrSessionServiceUnavailable
	}

	return models.WhatsappService.RestoreOrphaned()
}

// RestoreSessionManually exposes direct token-to-JID restore through the runtime layer.
func RestoreSessionManually(token string, jid string) error {
	if models.WhatsappService == nil {
		return ErrSessionServiceUnavailable
	}

	return models.WhatsappService.RestoreManual(token, jid)
}

// LoadSessionRecord materializes a stored session record into the live in-memory service cache.
func LoadSessionRecord(info *models.QpServer) (*models.QpWhatsappSession, error) {
	if info == nil {
		return nil, ErrNilSessionInfo
	}

	return models.WhatsappService.AppendNewSession(info)
}

// DeleteSessionRecord removes a session from persistence and tears down any live connection state.
func DeleteSessionRecord(session *models.QpWhatsappSession, cause string) error {
	if session == nil {
		return ErrNilSession
	}

	return models.WhatsappService.DeleteSession(session, cause)
}

// ApplySessionUser validates and applies a user change to the session without persisting it.
func ApplySessionUser(session *models.QpWhatsappSession, username string) (string, error) {
	if session == nil {
		return "", ErrNilSession
	}

	if len(username) == 0 || session.QpServer.GetUser() == username {
		return "", nil
	}

	if _, err := models.WhatsappService.DB.Users.Find(username); err != nil {
		return "", fmt.Errorf("user not found: %v", err.Error())
	}

	session.QpServer.SetUser(username)
	return fmt.Sprintf("user to: {%s}; ", username), nil
}

// ToggleSessionDebug flips the session debug mode and persists the new value.
func ToggleSessionDebug(session *models.QpWhatsappSession) (bool, error) {
	if session == nil {
		return false, ErrNilSession
	}

	return session.ToggleDevel()
}

// ToggleSessionOption flips one persisted session option and returns the new value.
func ToggleSessionOption(session *models.QpWhatsappSession, option string) (bool, error) {
	if session == nil {
		return false, ErrNilSession
	}

	switch option {
	case "groups":
		if err := models.ToggleGroups(session); err != nil {
			return false, err
		}
		return session.GetGroups(), nil
	case "broadcasts":
		if err := models.ToggleBroadcasts(session); err != nil {
			return false, err
		}
		return session.GetBroadcasts(), nil
	case "readreceipts":
		if err := models.ToggleReadReceipts(session); err != nil {
			return false, err
		}
		return session.GetReadReceipts(), nil
	case "calls":
		if err := models.ToggleCalls(session); err != nil {
			return false, err
		}
		return session.GetCalls(), nil
	case "readupdate":
		if err := models.ToggleReadUpdate(session); err != nil {
			return false, err
		}
		return session.ReadUpdate.Boolean(), nil
	default:
		return false, fmt.Errorf("unsupported option: %s", option)
	}
}

// ApplySessionConfigurationPatch mutates in-memory session flags and returns a human-readable update summary.
func ApplySessionConfigurationPatch(session *models.QpWhatsappSession, patch *SessionConfigurationPatch) (string, error) {
	if session == nil {
		return "", ErrNilSession
	}

	if patch == nil {
		return "", nil
	}

	update := ""

	if patch.Groups != nil && session.Groups != *patch.Groups {
		session.Groups = *patch.Groups
		update += fmt.Sprintf("groups to: {%s}; ", *patch.Groups)
	}

	if patch.Broadcasts != nil && session.Broadcasts != *patch.Broadcasts {
		session.Broadcasts = *patch.Broadcasts
		update += fmt.Sprintf("broadcasts to: {%s}; ", *patch.Broadcasts)
	}

	if patch.ReadReceipts != nil && session.ReadReceipts != *patch.ReadReceipts {
		session.ReadReceipts = *patch.ReadReceipts
		update += fmt.Sprintf("readreceipts to: {%s}; ", *patch.ReadReceipts)
	}

	if patch.Calls != nil && session.Calls != *patch.Calls {
		session.Calls = *patch.Calls
		update += fmt.Sprintf("calls to: {%s}; ", *patch.Calls)
	}

	if patch.ReadUpdate != nil && session.ReadUpdate != *patch.ReadUpdate {
		session.ReadUpdate = *patch.ReadUpdate
		update += fmt.Sprintf("readupdate to: {%s}; ", *patch.ReadUpdate)
	}

	if patch.Devel != nil && session.Devel != *patch.Devel {
		session.Devel = *patch.Devel
		update += fmt.Sprintf("devel to: {%t}; ", *patch.Devel)
	}

	return update, nil
}
