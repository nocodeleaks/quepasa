package models

import whatsapp "github.com/nocodeleaks/quepasa/whatsapp"

// QpWhatsappSession is the preferred domain name for the per-identity runtime
// object. It currently aliases the legacy server type to preserve compatibility
// while call sites migrate incrementally.
type QpWhatsappSession = QpWhatsappServer

// IQpWhatsappSession preserves interface compatibility during the naming
// migration from server to session.
type IQpWhatsappSession = IQpWhatsappServer

var ErrSessionNotFound = ErrServerNotFound

func PostToDispatchingFromSession(session *QpWhatsappSession, message *whatsapp.WhatsappMessage) error {
	return PostToDispatchingFromServer(session, message)
}

func PostToDispatchingsFromSession(session *QpWhatsappSession, dispatchings []*QpDispatching, message *whatsapp.WhatsappMessage) error {
	return PostToDispatchings(session, dispatchings, message)
}

func PostToWebhooksForSession(session *QpWhatsappSession, message *whatsapp.WhatsappMessage) error {
	return PostToWebhooksModern(session, message)
}

func GetSessionFromID(source string) (*QpWhatsappSession, error) {
	return GetServerFromID(source)
}

func GetSessionFromBot(source QPBot) (*QpWhatsappSession, error) {
	return GetServerFromBot(source)
}

func GetSessionFirstAvailable() (*QpWhatsappSession, error) {
	return GetServerFirstAvailable()
}

func GetSessionFromToken(token string) (*QpWhatsappSession, error) {
	return GetServerFromToken(token)
}

func GetSessionsForUserID(user string) map[string]*QpWhatsappSession {
	return WhatsappService.GetSessionsForUser(user)
}

func GetSessionsForUser(user *QpUser) map[string]*QpWhatsappSession {
	return GetSessionsForUserID(user.Username)
}

func (service *QPWhatsappService) AppendNewSession(info *QpServer) (*QpWhatsappSession, error) {
	return service.AppendNewServer(info)
}

func (service *QPWhatsappService) AppendPairedSession(paired *QpWhatsappPairing) (*QpWhatsappSession, error) {
	return service.AppendPaired(paired)
}

func (service *QPWhatsappService) NewQpWhatsappSession(info *QpServer) (*QpWhatsappSession, error) {
	return service.NewQpWhatsappServer(info)
}

func (service *QPWhatsappService) GetOrCreateSessionFromToken(token string) (*QpWhatsappSession, error) {
	return service.GetOrCreateServerFromToken(token)
}

func (service *QPWhatsappService) GetOrCreateSession(user string, wid string) (*QpWhatsappSession, error) {
	return service.GetOrCreateServer(user, wid)
}

func (service *QPWhatsappService) DeleteSession(session *QpWhatsappSession, cause string) error {
	return service.Delete(session, cause)
}

func (service *QPWhatsappService) GetSessionsForUser(username string) map[string]*QpWhatsappSession {
	return service.GetServersForUser(username)
}

func (service *QPWhatsappService) FindSessionByToken(token string) (*QpWhatsappSession, error) {
	return service.FindByToken(token)
}
