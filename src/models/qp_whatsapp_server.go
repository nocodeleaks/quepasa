package models

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	types "go.mau.fi/whatsmeow/types"

	"github.com/google/uuid"
	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type QpWhatsappServer struct {
	library.LogStruct // logging
	*QpServer
	QpDataWebhooks

	// should auto reconnect, false for qrcode scanner
	Reconnect bool `json:"reconnect"`

	connection     whatsapp.IWhatsappConnection `json:"-"`
	syncConnection *sync.Mutex                  `json:"-"` // Objeto de sinaleiro para evitar chamadas simultâneas a este objeto
	syncMessages   *sync.Mutex                  `json:"-"` // Objeto de sinaleiro para evitar chamadas simultâneas a este objeto

	//Battery        *WhatsAppBateryStatus        `json:"battery,omitempty"`

	StartTime time.Time `json:"starttime,omitempty"`

	Handler *QPWhatsappHandlers `json:"-"`
	WebHook *QPWebhookHandler   `json:"-"`

	// Stop request token
	StopRequested bool                   `json:"-"`
	db            QpDataServersInterface `json:"-"`
}

// custom log entry with fields: wid
func (source *QpWhatsappServer) GetLogger() *log.Entry {
	if source != nil && source.LogEntry != nil {
		return source.LogEntry
	}

	logentry := library.NewLogEntry(source)
	if source != nil {
		logentry = logentry.WithField(LogFields.WId, source.Wid)
		source.LogEntry = logentry
	}

	logentry.Level = log.ErrorLevel
	logentry.Infof("generating new log entry for %s, with level: %s", reflect.TypeOf(source), logentry.Level)

	return logentry
}

func (source *QpWhatsappServer) GetValidConnection() (whatsapp.IWhatsappConnection, error) {
	if source == nil || source.connection == nil || source.connection.IsInterfaceNil() {
		return nil, ErrorInvalidConnection
	}

	return source.connection, nil
}

//#region IMPLEMENTING WHATSAPP OPTIONS INTERFACE

func (source *QpWhatsappServer) GetOptions() *whatsapp.WhatsappOptions {
	if source == nil {
		return nil
	}

	return &source.WhatsappOptions
}

func (source *QpWhatsappServer) SetOptions(options *whatsapp.WhatsappOptions) error {
	source.WhatsappOptions = *options

	reason := fmt.Sprintf("options updated: %v", source.WhatsappOptions)
	return source.Save(reason)
}

//#endregion

// Ensure default handler
func (server *QpWhatsappServer) HandlerEnsure() {
	if server == nil {
		return // invalid state
	}

	if server.Handler == nil {
		handler := &QPWhatsappHandlers{
			server:       server,
			syncRegister: &sync.Mutex{},
		}

		logentry := server.GetLogger()
		logentry.Debug("ensuring messages handler for server")

		// logging
		handler.LogEntry = logentry

		// updating
		server.Handler = handler
	}
}

func (server *QpWhatsappServer) HasSignalRActiveConnections() bool {
	if server == nil {
		return false // invalid state
	}

	return SignalRHub.HasActiveConnections(server.Token)
}

//region IMPLEMENT OF INTERFACE STATE RECOVERY

func (server *QpWhatsappServer) GetStatus() whatsapp.WhatsappConnectionState {
	if server == nil {
		return whatsapp.Unknown // invalid state
	}

	if server.connection == nil {
		if server.Verified {
			if server.StopRequested {
				return whatsapp.Stopped
			}
			return whatsapp.UnPrepared
		}

		return whatsapp.UnVerified
	} else {
		if server.StopRequested {
			if server.connection != nil && !server.connection.IsInterfaceNil() && server.connection.IsConnected() {
				return whatsapp.Stopping
			} else {
				return whatsapp.Stopped
			}
		} else {
			state := server.connection.GetStatus()
			if state == whatsapp.Disconnected && !server.Verified {
				return whatsapp.UnVerified
			}
			return state
		}
	}
}

//endregion
//region IMPLEMENT OF INTERFACE QUEPASA SERVER

// Returns whatsapp controller id on E164
// Ex: 5521967609095
func (server QpWhatsappServer) GetWId() string {
	return server.QpServer.Wid
}

func (source *QpWhatsappServer) DownloadData(id string) ([]byte, error) {
	msg, err := source.Handler.GetById(id)
	if err != nil {
		return nil, err
	}

	source.GetLogger().Infof("downloading msg data %s", id)
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

	source.GetLogger().Infof("downloading msg %s, using cache: %v", id, cache)
	att, err = source.connection.Download(msg, cache)
	if err != nil {
		return
	}

	return
}

func (source *QpWhatsappServer) RevokeByPrefix(id string) (errors []error) {
	messages := source.Handler.GetByPrefix(id)
	for _, msg := range messages {
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

	source.GetLogger().Infof("revoking msg %s", id)
	return source.connection.Revoke(msg)
}

//endregion

//#region WEBHOOKS

func (source *QpWhatsappServer) GetWebHook(url string) *QpWhatsappServerWebhook {
	for _, item := range source.Webhooks {
		if item.Url == url {
			return &QpWhatsappServerWebhook{
				QpWebhook: item,
				server:    source,
			}
		}
	}
	return nil
}

func (source *QpWhatsappServer) GetWebHooksByUrl(filter string) (out []*QpWebhook) {
	for _, element := range source.Webhooks {
		if strings.Contains(element.Url, filter) {
			out = append(out, element)
		}
	}
	return
}

// Ensure default webhook handler
func (server *QpWhatsappServer) WebHookEnsure() {
	if server.WebHook == nil {
		webHookHandler := &QPWebhookHandler{server: server}

		logentry := server.GetLogger()
		logentry.Debug("ensuring webhook handler for server")

		// logging
		webHookHandler.LogEntry = logentry

		// updating
		server.WebHook = webHookHandler
	}
}

//#endregion

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

// Roda de forma assíncrona, não interessa o resultado ao chamador
// Inicia o processo de tentativas de conexão de um servidor individual
func (source *QpWhatsappServer) Initialize() {
	if source == nil {
		panic("nil server, code error")
	}

	logentry := source.GetLogger()
	logentry.Info("initializing whatsapp server ...")

	err := source.Start()
	if err != nil {
		logentry.Errorf("initializing server error: %s", err.Error())
	}
}

// Update underlying connection and ensure trivials
func (source *QpWhatsappServer) UpdateConnection(connection whatsapp.IWhatsappConnection) {

	if source.connection != nil && !source.connection.IsInterfaceNil() {
		source.connection.Dispose("UpdateConnection")
	}

	source.connection = connection
	if source.Handler == nil {
		logentry := source.GetLogger()
		logentry.Warn("creating handlers ?! not implemented yet")
	}

	source.connection.UpdateHandler(source.Handler)

	// Registrando webhook
	webhookDispatcher := &QPWebhookHandler{server: source}
	if !source.Handler.IsAttached() {
		source.Handler.Register(webhookDispatcher)
	}
}

func (source *QpWhatsappServer) EnsureUnderlying() (err error) {

	if len(source.Wid) > 0 && !source.Verified {
		err = fmt.Errorf("not verified")
		return
	}

	source.syncConnection.Lock()
	defer source.syncConnection.Unlock()

	// conectar dispositivo
	if source.connection == nil {
		logentry := source.GetLogger()

		options := &whatsapp.WhatsappConnectionOptions{
			WhatsappOptions: &source.WhatsappOptions,
			Wid:             source.Wid,
			Reconnect:       true,
			LogStruct:       library.LogStruct{LogEntry: logentry},
		}

		logentry.Infof("trying to create new whatsapp connection, auto reconnect: %v, log level: %s", options.Reconnect, logentry.Level)

		connection, err := NewConnection(options)
		if err != nil {
			waError, ok := err.(whatsapp.WhatsappError)
			if ok {
				if waError.Unauthorized() {
					source.MarkVerified(false)
				}
			}
			return err
		} else {
			source.connection = connection
		}
	}

	return
}

// called from service started, after retrieve servers from database
func (source *QpWhatsappServer) Start() (err error) {
	logentry := source.GetLogger()

	logentry.Infof("starting whatsapp server, with log level: %s", logentry.Level)
	err = source.EnsureUnderlying()
	if err != nil {
		return
	}

	state := source.GetStatus()
	logentry.Debugf("starting whatsapp server ... on %s state", state)

	if !IsValidToStart(state) {
		err = fmt.Errorf("trying to start a server on an invalid state :: %s", state)
		logentry.Warnf(err.Error())
		return
	}

	// reset stop requested token
	source.StopRequested = false

	if !source.Handler.IsAttached() {

		// Registrando webhook
		source.Handler.Register(source.WebHook)
	}

	// Atualizando manipuladores de eventos
	source.connection.UpdateHandler(source.Handler)

	logentry.Infof("requesting connection ...")
	err = source.connection.Connect()
	if err != nil {
		return source.StartConnectionError(err)
	}

	if !source.connection.IsConnected() {
		logentry.Infof("requesting connection again ...")
		err = source.connection.Connect()
		if err != nil {
			return source.StartConnectionError(err)
		}
	}

	// If at this moment the connect is already logged, ensure a valid mark
	if source.connection.IsValid() {
		source.MarkVerified(true)
	}

	return
}

// called after success paring devices
func (source *QpWhatsappServer) EnsureReady() (err error) {
	logger := source.GetLogger()

	logger.Info("ensuring that whatsapp server is ready")
	err = source.EnsureUnderlying()
	if err != nil {
		logger.Errorf("error on ensure underlaying connection: %s", err.Error())
		return
	}

	// reset stop requested token
	source.StopRequested = false

	if !source.Handler.IsAttached() {
		logger.Info("attaching handlers")

		// Registrando webhook
		source.Handler.Register(source.WebHook)
	} else {
		logger.Debug("handlers already attached")
	}

	// Atualizando manipuladores de eventos
	source.connection.UpdateHandler(source.Handler)

	if !source.connection.IsConnected() {
		logger.Info("requesting connection ...")
		err = source.connection.Connect()
		if err != nil {
			return source.StartConnectionError(err)
		}
	} else {
		logger.Debug("already connected")
	}

	// If at this moment the connect is already logged, ensure a valid mark
	source.MarkVerified(true)

	return
}

// Process an error at start connection
func (source *QpWhatsappServer) StartConnectionError(err error) error {
	logger := source.GetLogger()

	source.Disconnect("StartConnectionError")
	source.Handler.Clear()

	if _, ok := err.(*whatsapp.UnAuthorizedError); ok {
		logger.Warningf("unauthorized, setting unverified")
		return source.MarkVerified(false)
	}

	logger.Errorf("error on starting whatsapp server connection: %s", err.Error())
	return err
}

func (source *QpWhatsappServer) Stop(cause string) (err error) {
	if !source.StopRequested {

		// setting token
		source.StopRequested = true

		// loggging properly
		logentry := source.GetLogger()
		logentry.Infof("stopping server: %s", cause)

		source.Disconnect("stop: " + cause)

		if source.Handler != nil {
			source.Handler.Clear()
		}
	}

	return
}

func (source *QpWhatsappServer) Restart() (err error) {
	err = source.Stop("restart")
	if err != nil {
		return
	}

	// wait 1 second before continue
	time.Sleep(1 * time.Second)

	logentry := source.GetLogger()
	logentry.Info("re-initializing whatsapp server ...")

	return source.Start()
}

// Somente usar em caso de não ser permitida a reconxão automática
func (source *QpWhatsappServer) Disconnect(cause string) {
	conn, err := source.GetValidConnection()
	if err == nil {
		if conn.IsConnected() {
			logentry := source.GetLogger()
			logentry.Infof("disconnecting whatsapp server by: %s", cause)

			conn.Dispose(cause)
			source.connection = nil
		}
	}
}

// Retorna o titulo em cache (se houver) do id passado em parametro
func (source *QpWhatsappServer) GetChatTitle(wid string) string {
	conn, err := source.GetValidConnection()
	if err != nil {
		return ""
	}

	return conn.GetChatTitle(wid)
}

// Usado para exibir os servidores/bots de cada usuario em suas respectivas telas
func (server *QpWhatsappServer) GetOwnerID() string {
	return server.User
}

//region QP BOT EXTENSIONS

// Check if the current connection state is valid for a start method
func IsValidToStart(status whatsapp.WhatsappConnectionState) bool {
	if status == whatsapp.Stopped {
		return true
	}
	if status == whatsapp.Stopping {
		return true
	}
	if status == whatsapp.Disconnected {
		return true
	}
	if status == whatsapp.Failed {
		return true
	}
	return false
}

func (source *QpWhatsappServer) GetWorking() bool {
	status := source.GetStatus()
	return !IsValidToStart(status)
}

func (server *QpWhatsappServer) GetStatusString() string {
	return server.GetStatus().String()
}

func (server *QpWhatsappServer) ID() string {
	return server.Wid
}

// Traduz o Wid para um número de telefone em formato E164
func (server *QpWhatsappServer) GetNumber() string {
	return library.GetPhoneByWId(server.Wid)
}

func (server *QpWhatsappServer) GetTimestamp() time.Time {
	return server.Timestamp
}

func (server *QpWhatsappServer) GetStartedTime() time.Time {
	return server.StartTime
}

func (server *QpWhatsappServer) GetConnection() whatsapp.IWhatsappConnection {
	return server.connection
}

func (server *QpWhatsappServer) Toggle() (err error) {
	if !server.GetWorking() {
		err = server.Start()
	} else {
		err = server.Stop("toggling")
	}
	return
}

func (server *QpWhatsappServer) IsDevelopmentGlobal() bool {
	switch ENV.LogLevel() {
	case "debug", "trace":
		return true
	default:
		return false
	}
}

/*
<summary>

	Set a new random Guid token for whatsapp server bot

</summary>
*/
func (server *QpWhatsappServer) CycleToken() (err error) {
	value := uuid.New().String()
	return server.UpdateToken(value)
}

/*
<summary>

	Set a specific not empty token for whatsapp server bot

</summary>
*/
func (source *QpWhatsappServer) UpdateToken(value string) (err error) {
	if len(value) == 0 {
		err = fmt.Errorf("empty token")
		return
	}

	err = source.UpdateToken(value)
	if err != nil {
		return
	}

	source.GetLogger().Infof("updating token: %v", value)
	return
}

/*
<summary>

	Get current token for whatsapp server bot

</summary>
*/
func (server *QpWhatsappServer) GetToken() string {
	return server.Token
}

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
	source.Timestamp = time.Now().UTC()

	if ok {
		logger.Debugf("updating server info: %+v", source)
		return source.db.Update(source.QpServer)
	} else {
		logger.Debugf("adding server info: %+v", source)
		return source.db.Add(source.QpServer)
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

//endregion

// delete this whatsapp server and underlaying connection
func (server *QpWhatsappServer) Delete() error {
	if server.connection != nil {
		err := server.connection.Delete()
		if err != nil {
			return fmt.Errorf("whatsapp server, delete connection, error: %s", err.Error())
		}

		server.connection = nil
	}

	err := server.QpDataWebhooks.WebhookClear()
	if err != nil {
		return fmt.Errorf("whatsapp server, webhook clear, error: %s", err.Error())
	}

	err = server.db.Delete(server.Token)
	if err != nil {
		return fmt.Errorf("whatsapp server, database delete connection, error: %s", err.Error())
	}

	return nil
}

//endregion
//#region SEND

// Default send message method
func (source *QpWhatsappServer) SendMessage(msg *whatsapp.WhatsappMessage) (response whatsapp.IWhatsappSendResponse, err error) {
	logger := source.GetLogger()
	logger.Debugf("sending msg to: %s", msg.Chat.Id)

	conn, err := source.GetValidConnection()
	if err != nil {
		return
	}

	// leading with wrongs digit 9
	if ENV.ShouldRemoveDigit9() {

		phone, _ := library.ExtractPhoneIfValid(msg.Chat.Id)
		if len(phone) > 0 {
			phoneWithout9, _ := library.RemoveDigit9IfElegible(phone)
			if len(phoneWithout9) > 0 {
				valids, err := conn.IsOnWhatsApp(phone, phoneWithout9)
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

//#endregion
//#region PROFILE PICTURE

func (source *QpWhatsappServer) GetProfilePicture(wid string, knowingId string) (picture *whatsapp.WhatsappProfilePicture, err error) {
	logger := source.GetLogger()
	logger.Debugf("getting info about profile picture for: %s, with id: %s", wid, knowingId)

	// future implement a rate control here, high volume of requests causing bans
	// studying rates ...

	conn, err := source.GetValidConnection()
	if err != nil {
		return
	}

	return conn.GetProfilePicture(wid, knowingId)
}

//#endregion
//#region GROUP INVITE LINK

func (source *QpWhatsappServer) GetInvite(groupId string) (link string, err error) {
	conn, err := source.GetValidConnection()
	if err != nil {
		return
	}

	return conn.GetInvite(groupId)
}

//#endregion
//#region GET ALL CONTACTS

func (source *QpWhatsappServer) GetContacts() (contacts []whatsapp.WhatsappChat, err error) {
	conn, err := source.GetValidConnection()
	if err != nil {
		return
	}

	contacts, err = conn.GetContacts()
	if err == nil {
		for index, contact := range contacts {
			contact.FormatContact()
			contacts[index] = contact
		}
	}

	return
}

//#endregion

//#region IsOnWhatsapp

func (source *QpWhatsappServer) IsOnWhatsApp(phones ...string) (registered []string, err error) {
	conn, err := source.GetValidConnection()
	if err != nil {
		return
	}

	return conn.IsOnWhatsApp(phones...)
}

//#endregion

// #region GROUPS
func (server *QpWhatsappServer) GetJoinedGroups() ([]*types.GroupInfo, error) {
	conn, err := server.GetValidConnection()
	if err != nil {
		return nil, err
	}

	return conn.GetJoinedGroups()
}

func (server *QpWhatsappServer) GetGroupInfo(groupID string) (*types.GroupInfo, error) {
	conn, err := server.GetValidConnection()
	if err != nil {
		return nil, err
	}

	return conn.GetGroupInfo(groupID)
}

func (server *QpWhatsappServer) CreateGroup(name string, participants []string) (*types.GroupInfo, error) {
	conn, err := server.GetValidConnection()
	if err != nil {
		return nil, err
	}

	return conn.CreateGroup(name, participants)
}

func (server *QpWhatsappServer) UpdateGroupSubject(groupID string, name string) (*types.GroupInfo, error) {
	conn, err := server.GetValidConnection() // Ensure a valid connection is available
	if err != nil {
		return nil, err
	}

	return conn.UpdateGroupSubject(groupID, name) // Delegate the call to the connection
}

func (server *QpWhatsappServer) UpdateGroupPhoto(groupID string, imageData []byte) (string, error) {
	conn, err := server.GetValidConnection()
	if err != nil {
		return "", err
	}

	return conn.UpdateGroupPhoto(groupID, imageData)
}

func (server *QpWhatsappServer) UpdateGroupParticipants(groupJID string, participants []string, action string) ([]interface{}, error) {
	conn, err := server.GetValidConnection()
	if err != nil {
		return nil, err
	}

	return conn.UpdateGroupParticipants(groupJID, participants, action)
}

func (server *QpWhatsappServer) GetGroupJoinRequests(groupJID string) ([]interface{}, error) {
	conn, err := server.GetValidConnection()
	if err != nil {
		return nil, err
	}

	return conn.GetGroupJoinRequests(groupJID)
}

func (server *QpWhatsappServer) HandleGroupJoinRequests(groupJID string, participants []string, action string) ([]interface{}, error) {
	conn, err := server.GetValidConnection()
	if err != nil {
		return nil, err
	}

	return conn.HandleGroupJoinRequests(groupJID, participants, action)
}

func (server *QpWhatsappServer) CreateGroupExtended(options map[string]interface{}) (*types.GroupInfo, error) {
	conn, err := server.GetValidConnection()
	if err != nil {
		return nil, err
	}

	// Extract parameters
	title, _ := options["title"].(string)
	participantsRaw, _ := options["participants"].([]string)

	return conn.CreateGroupExtended(title, participantsRaw)
}

// Add to QpWhatsappServer
func (server *QpWhatsappServer) SendChatPresence(chatId string, isTyping bool, mediaType string) error {
	conn, err := server.GetValidConnection()
	if err != nil {
		return err
	}
	return conn.SendChatPresence(chatId, isTyping, mediaType)
}

//#endregion
