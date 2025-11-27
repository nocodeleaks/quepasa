package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	library "github.com/nocodeleaks/quepasa/library"
	rabbitmq "github.com/nocodeleaks/quepasa/rabbitmq"
	signalr "github.com/nocodeleaks/quepasa/signalr"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type QpWhatsappServer struct {
	*QpServer
	QpDataDispatching // new dispatching system

	// should auto reconnect, false for qrcode scanner
	Reconnect bool `json:"reconnect"`

	connection     whatsapp.IWhatsappConnection `json:"-"`
	syncConnection *sync.Mutex                  `json:"-"` // Objeto de sinaleiro para evitar chamadas simultâneas a este objeto
	syncMessages   *sync.Mutex                  `json:"-"` // Objeto de sinaleiro para evitar chamadas simultâneas a este objeto

	Timestamps QpTimestamps `json:"timestamps"`

	Handler            *DispatchingHandler   `json:"-"`
	DispatchingHandler *QPDispatchingHandler `json:"-"`
	GroupManager       *QpGroupManager       `json:"-"` // composition for group operations
	StatusManager      *QpStatusManager      `json:"-"` // composition for status operations
	ContactManager     *QpContactManager     `json:"-"` // composition for contact operations

	// Stop request token
	StopRequested bool                   `json:"-"`
	db            QpDataServersInterface `json:"-"`
}

// MarshalJSON customizes JSON serialization to include only dispatching field instead of webhooks
func (source QpWhatsappServer) MarshalJSON() ([]byte, error) {
	// Create a custom struct to control serialization
	type CustomServer struct {
		*QpServer
		Reconnect   bool             `json:"reconnect"`
		StartTime   time.Time        `json:"starttime,omitempty"`
		Timestamps  QpTimestamps     `json:"timestamps"`
		Dispatching []*QpDispatching `json:"dispatching,omitempty"`
	}

	// Get dispatching data from memory (includes real-time failure/success updates)
	var dispatchingData []*QpDispatching
	if source.QpDataDispatching.Dispatching != nil {
		// Use in-memory dispatching data with real-time status
		dispatchingData = source.QpDataDispatching.Dispatching
	}

	// Prepare timestamps for serialization
	timestamps := source.Timestamps
	timestamps.Update = source.Timestamp

	custom := CustomServer{
		QpServer:    source.QpServer,
		Reconnect:   source.Reconnect,
		StartTime:   timestamps.Start,
		Timestamps:  timestamps,
		Dispatching: dispatchingData,
	}

	return json.Marshal(custom)
}

func (source *QpWhatsappServer) GetValidConnection() (whatsapp.IWhatsappConnection, error) {
	if source == nil || source.connection == nil || source.connection.IsInterfaceNil() {
		return nil, ErrInvalidConnection
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
		handler := &DispatchingHandler{
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

	return signalr.SignalRHub.HasActiveConnections(server.Token)
}

//region IMPLEMENT OF INTERFACE STATE RECOVERY

func (server *QpWhatsappServer) GetStatus() whatsapp.WhatsappConnectionState {
	return server.GetState()
}

// GetState retrieves the current calculated connection state of the WhatsApp server
func (server *QpWhatsappServer) GetState() whatsapp.WhatsappConnectionState {
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
			statusManager := server.GetStatusManager()
			if server.connection != nil && !server.connection.IsInterfaceNil() && statusManager.IsConnected() {
				return whatsapp.Stopping
			} else {
				return whatsapp.Stopped
			}
		} else {
			statusManager := server.GetStatusManager()
			state := statusManager.GetState()
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

//endregion

//#region DEPRECATED LEGACY METHODS (TO BE REMOVED)

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
	} else {
		// Update handler on the new connection
		source.connection.UpdateHandler(source.Handler)
	}

	// Registrando webhook
	dispatchingHandler := &QPDispatchingHandler{server: source}
	if !source.Handler.IsAttached() {
		source.Handler.Register(dispatchingHandler)
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
			ExternalHandler: source.Handler, // Pass handler to connection
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
			// Handler is already configured via ExternalHandler in options
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
		logentry.Warnf("%s", err.Error())
		return
	}

	// reset stop requested token
	source.StopRequested = false

	// Update start timestamp
	source.Timestamps.Start = time.Now().UTC()

	if !source.Handler.IsAttached() {

		// Registrando dispatching handler
		source.Handler.Register(source.DispatchingHandler)
	}

	// Handler already configured during connection creation via ExternalHandler
	// No need to call UpdateHandler here

	// Initialize RabbitMQ connections for this server
	source.InitializeRabbitMQConnections()

	logentry.Infof("requesting connection ...")
	err = source.connection.Connect()
	if err != nil {
		return source.StartConnectionError(err)
	}

	statusManager := source.GetStatusManager()
	if !statusManager.IsConnected() {
		logentry.Infof("requesting connection again ...")
		err = source.connection.Connect()
		if err != nil {
			return source.StartConnectionError(err)
		}
	}

	// If at this moment the connect is already logged, ensure a valid mark
	if statusManager.IsValid() {
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

		// Registrando dispatching handler
		source.Handler.Register(source.DispatchingHandler)
	} else {
		logger.Debug("handlers already attached")
	}

	// Handler already configured during connection creation via ExternalHandler
	// No need to call UpdateHandler here

	statusManager := source.GetStatusManager()
	if !statusManager.IsConnected() {
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
		statusManager := source.GetStatusManager()
		if statusManager.IsConnected() {
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
	return server.Timestamps.Update
}

func (server *QpWhatsappServer) GetStartedTime() time.Time {
	return server.Timestamps.Start
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
	currentTime := time.Now().UTC()
	source.Timestamp = currentTime
	source.Timestamps.Update = currentTime

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

	// Clear dispatching data from new system
	db := GetDatabase()
	if db != nil && db.Dispatching != nil {
		err := db.Dispatching.Clear(server.Token)
		if err != nil {
			return fmt.Errorf("whatsapp server, dispatching clear, error: %s", err.Error())
		}
	}

	err := server.db.Delete(server.Token)
	if err != nil {
		return fmt.Errorf("whatsapp server, database delete connection, error: %s", err.Error())
	}

	return nil
}

//#endregion
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

		phone, _ := whatsapp.GetPhoneIfValid(msg.Chat.Id)
		if len(phone) > 0 {
			phoneWithout9, _ := library.RemoveDigit9IfElegible(phone)
			if len(phoneWithout9) > 0 {
				contactManager := source.GetContactManager()
				valids, err := contactManager.IsOnWhatsApp(phone, phoneWithout9)
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

	contactManager := source.GetContactManager()
	return contactManager.GetProfilePicture(wid, knowingId)
}

//#endregion
//#region GROUP INVITE LINK

//#endregion
//#region GET ALL CONTACTS

func (source *QpWhatsappServer) GetContacts() (contacts []whatsapp.WhatsappChat, err error) {
	contactManager := source.GetContactManager()
	contacts, err = contactManager.GetContacts()
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
	contactManager := source.GetContactManager()
	return contactManager.IsOnWhatsApp(phones...)
}

//#endregion

// #region GROUPS

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

//#endregion

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

//#endregion

//#region RABBITMQ CONFIGS

func (source *QpWhatsappServer) GetRabbitMQConfig(exchangeName string) *QpRabbitMQConfig {
	configs := source.QpDataDispatching.GetRabbitMQConfigs()
	for _, config := range configs {
		if config.ExchangeName == exchangeName {
			return config
		}
	}
	return nil
}

func (source *QpWhatsappServer) GetRabbitMQConfigsByQueue(filter string) (out []*QpRabbitMQConfig) {
	configs := source.QpDataDispatching.GetRabbitMQConfigs()
	for _, element := range configs {
		if len(filter) == 0 || strings.Contains(element.ExchangeName, filter) {
			out = append(out, element)
		}
	}
	return
}

// GetRabbitMQConfigs returns all RabbitMQ configurations for this server
func (source *QpWhatsappServer) GetRabbitMQConfigs() []*QpRabbitMQConfig {
	db := GetDatabase()
	if db != nil && db.Dispatching != nil {
		dispatchings, err := db.Dispatching.FindAll(source.Token)
		if err == nil {
			var configs []*QpRabbitMQConfig
			for _, dispatching := range dispatchings {
				if dispatching.QpDispatching != nil && dispatching.Type == DispatchingTypeRabbitMQ {
					config := &QpRabbitMQConfig{
						ConnectionString: dispatching.ConnectionString,
						TrackId:          dispatching.TrackId,
						ForwardInternal:  dispatching.ForwardInternal,
						Extra:            dispatching.Extra,
						Timestamp:        dispatching.Timestamp,
					}
					configs = append(configs, config)
				}
			}
			return configs
		}
	}
	return []*QpRabbitMQConfig{}
}

// HasRabbitMQConfigs returns true if the server has RabbitMQ configurations
func (server *QpWhatsappServer) HasRabbitMQConfigs() bool {
	configs := server.GetRabbitMQConfigsByQueue("")
	return len(configs) > 0
}

// HasWebhooks returns true if the server has webhook configurations
func (server *QpWhatsappServer) HasWebhooks() bool {
	webhooks := server.GetWebhooks()
	return len(webhooks) > 0
}

//#endregion

//#region DISPATCHING

// Get dispatching by connection string
func (source *QpWhatsappServer) GetDispatching(connectionString string) *QpDispatching {
	db := GetDatabase()
	if db != nil && db.Dispatching != nil {
		dispatching, err := db.Dispatching.Find(source.Token, connectionString)
		if err == nil && dispatching != nil {
			return dispatching.QpDispatching
		}
	}
	return nil
}

// Get dispatching by connection string and type
func (source *QpWhatsappServer) GetDispatchingByType(connectionString string, dispatchType string) *QpDispatching {
	for _, item := range source.QpDataDispatching.Dispatching {
		if item.ConnectionString == connectionString && item.Type == dispatchType {
			return item
		}
	}
	return nil
}

// Get all dispatching by filter
func (source *QpWhatsappServer) GetDispatchingByFilter(filter string) (out []*QpDispatching) {
	for _, element := range source.QpDataDispatching.Dispatching {
		if len(filter) == 0 || strings.Contains(element.ConnectionString, filter) {
			out = append(out, element)
		}
	}
	return
}

// Ensure dispatching handler
func (server *QpWhatsappServer) DispatchingEnsure() {
	if server.DispatchingHandler == nil {
		dispatchingHandler := &QPDispatchingHandler{server: server}

		logentry := server.GetLogger()
		logentry.Debug("ensuring dispatching handler for server")

		// logging
		dispatchingHandler.LogEntry = logentry

		// updating
		server.DispatchingHandler = dispatchingHandler
	}
}

// GetWebhookDispatchings returns all webhook configurations as QpDispatching
func (source *QpWhatsappServer) GetWebhookDispatchings() []*QpDispatching {
	allDispatchings := source.GetDispatchingByFilter("")
	webhooks := []*QpDispatching{}

	for _, dispatching := range allDispatchings {
		if dispatching.IsWebhook() {
			webhooks = append(webhooks, dispatching)
		}
	}

	return webhooks
}

// GetWebhooks returns webhook dispatchings converted to QpWebhook format for interface compatibility
func (source *QpWhatsappServer) GetWebhooks() []*QpWebhook {
	return source.QpDataDispatching.GetWebhooks()
}

// InitializeRabbitMQConnections initializes all RabbitMQ connections for this server
func (source *QpWhatsappServer) InitializeRabbitMQConnections() {
	logentry := source.GetLogger()

	// Get all RabbitMQ configurations for this server
	configs := source.GetRabbitMQConfigs()

	if len(configs) == 0 {
		logentry.Debug("no RabbitMQ configurations found for this server")
		return
	}

	logentry.Infof("initializing %d RabbitMQ connection(s) for server", len(configs))

	for _, config := range configs {
		if config.ConnectionString != "" {
			logentry.Infof("initializing RabbitMQ connection: %s", config.ConnectionString)

			// Import rabbitmq package and get client to initialize connection
			// This will create the connection pool if it doesn't exist
			client := rabbitmq.GetRabbitMQClient(config.ConnectionString)
			if client != nil {
				logentry.Infof("RabbitMQ connection initialized successfully: %s", config.ConnectionString)
			} else {
				logentry.Warnf("failed to initialize RabbitMQ connection: %s", config.ConnectionString)
			}
		}
	}
}

//#endregion
