package models

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/google/uuid"
	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type QpWhatsappServer struct {
	*QpServer
	QpDataWebhooks
	connection     whatsapp.IWhatsappConnection `json:"-"`
	syncConnection *sync.Mutex                  `json:"-"` // Objeto de sinaleiro para evitar chamadas simultâneas a este objeto
	syncMessages   *sync.Mutex                  `json:"-"` // Objeto de sinaleiro para evitar chamadas simultâneas a este objeto
	Battery        *WhatsAppBateryStatus        `json:"battery,omitempty"`
	StartTime      time.Time                    `json:"starttime,omitempty"`
	Handler        *QPWhatsappHandlers          `json:"-"`
	WebHook        *QPWebhookHandler            `json:"-"`

	stopRequested bool                   `json:"-"`
	Log           *log.Entry             `json:"-"`
	db            QpDataServersInterface `json:"-"`
}

// Ensure default handler
func (server *QpWhatsappServer) HandlerEnsure() {
	if server.Handler == nil {

		if server.Log == nil {
			server.Log = log.NewEntry(log.StandardLogger())
		}

		handlerMessages := make(map[string]whatsapp.WhatsappMessage)
		handler := &QPWhatsappHandlers{
			server:       server,
			messages:     handlerMessages,
			sync:         &sync.Mutex{},
			syncRegister: &sync.Mutex{},
		}

		server.Handler = handler
	}
}

// Ensure default webhook handler
func (server *QpWhatsappServer) WebHookEnsure() {
	if server.WebHook == nil {
		server.WebHook = &QPWebhookHandler{server}
	}
}

//endregion
//region IMPLEMENT OF INTERFACE STATE RECOVERY

func (server *QpWhatsappServer) GetStatus() whatsapp.WhatsappConnectionState {
	if server.connection == nil {
		if server.Verified {
			return whatsapp.UnPrepared
		} else {
			return whatsapp.UnVerified
		}
	} else {
		state := server.connection.GetStatus()
		if server.stopRequested && !server.connection.IsConnected() {
			return whatsapp.Stopped
		} else {
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
func (server *QpWhatsappServer) GetWid() string {
	return server.WId
}

func (server *QpWhatsappServer) DownloadData(id string) ([]byte, error) {
	msg, err := server.Handler.GetMessage(id)
	if err != nil {
		return nil, err
	}

	server.Log.Infof("downloading msg data %s", id)
	return server.connection.DownloadData(&msg)
}

/*
<summary>

	Download attachment from msg id, optional use cached data or not

</summary>
*/
func (server *QpWhatsappServer) Download(id string, cache bool) (att *whatsapp.WhatsappAttachment, err error) {
	msg, err := server.Handler.GetMessage(id)
	if err != nil {
		return
	}

	server.Log.Infof("downloading msg %s, using cache: %v", id, cache)
	att, err = server.connection.Download(&msg, cache)
	if err != nil {
		return
	}

	return
}

func (server *QpWhatsappServer) Revoke(id string) (err error) {
	msg, err := server.Handler.GetMessage(id)
	if err != nil {
		return
	}

	server.Log.Infof("revoking msg %s", id)
	return server.connection.Revoke(&msg)
}

//endregion

func (server *QpWhatsappServer) GetMessages(timestamp time.Time) (messages []whatsapp.WhatsappMessage) {
	messages = append(messages, server.Handler.GetMessages(timestamp)...)
	return
}

// Roda de forma assíncrona, não interessa o resultado ao chamador
// Inicia o processo de tentativas de conexão de um servidor individual
func (server *QpWhatsappServer) Initialize() {
	if server == nil {
		panic("nil server, code error")
	}

	server.Log.Info("initializing whatsapp server ...")
	err := server.Start()
	if err != nil {
		server.Log.Errorf("initializing server error: %s", err.Error())
	}
}

// Update underlying connection and ensure trivials
func (server *QpWhatsappServer) UpdateConnection(connection whatsapp.IWhatsappConnection) {

	if server.connection != nil && !server.connection.IsInterfaceNil() {
		server.connection.Dispose("UpdateConnection")
	}

	server.connection = connection
	server.connection.UpdateLog(server.Log)
	if server.Handler == nil {
		server.Log.Info("creating handlers ?!")
	}

	server.connection.UpdateHandler(server.Handler)

	// Registrando webhook
	webhookDispatcher := &QPWebhookHandler{server}
	if !server.Handler.IsAttached() {
		server.Handler.Register(webhookDispatcher)
	}
}

func (server *QpWhatsappServer) EnsureUnderlying() (err error) {

	if len(server.WId) > 0 && !server.Verified {
		err = fmt.Errorf("not verified")
		return
	}

	server.syncConnection.Lock()
	defer server.syncConnection.Unlock()

	// conectar dispositivo
	if server.connection == nil {

		server.Log.Infof("trying to create new whatsapp connection ...")
		connection, err := NewConnection(server.WId, server.Log)
		if err != nil {
			waError, ok := err.(whatsapp.WhatsappError)
			if ok {
				if waError.Unauthorized() {
					server.MarkVerified(false)
				}
			}
			return err
		} else {
			server.connection = connection
		}
	}

	return
}

func (server *QpWhatsappServer) Start() (err error) {
	server.Log.Info("starting whatsapp server")
	err = server.EnsureUnderlying()
	if err != nil {
		return
	}

	state := server.GetStatus()
	server.Log.Debugf("starting whatsapp server ... on %s state", state)

	if !IsValidToStart(state) {
		err = fmt.Errorf("trying to start a server on an invalid state :: %s", state)
		server.Log.Warnf(err.Error())
		return
	}

	// reset stop requested token
	server.stopRequested = false

	if !server.Handler.IsAttached() {

		// Registrando webhook
		server.Handler.Register(server.WebHook)
	}

	// Atualizando manipuladores de eventos
	server.connection.UpdateHandler(server.Handler)

	server.Log.Infof("requesting connection ...")
	err = server.connection.Connect()
	if err != nil {
		return server.StartConnectionError(err)
	}

	if !server.connection.IsConnected() {
		server.Log.Infof("requesting connection again ...")
		err = server.connection.Connect()
		if err != nil {
			return server.StartConnectionError(err)
		}
	}

	// If at this moment the connect is already logged, ensure a valid mark
	if server.connection.IsValid() {
		server.MarkVerified(true)
	}

	return
}

func (server *QpWhatsappServer) EnsureReady() (err error) {
	server.Log.Info("ensuring that whatsapp server is ready")
	err = server.EnsureUnderlying()
	if err != nil {
		server.Log.Errorf("error on ensure underlaying connection: %s", err.Error())
		return
	}

	// reset stop requested token
	server.stopRequested = false

	if !server.Handler.IsAttached() {
		server.Log.Info("attaching handlers")

		// Registrando webhook
		server.Handler.Register(server.WebHook)
	} else {
		server.Log.Debug("handlers already attached")
	}

	// Atualizando manipuladores de eventos
	server.connection.UpdateHandler(server.Handler)

	if !server.connection.IsConnected() {
		server.Log.Info("requesting connection ...")
		err = server.connection.Connect()
		if err != nil {
			return server.StartConnectionError(err)
		}
	} else {
		server.Log.Debug("already connected")
	}

	// If at this moment the connect is already logged, ensure a valid mark
	server.MarkVerified(true)

	return
}

// Process an error at start connection
func (server *QpWhatsappServer) StartConnectionError(err error) error {
	server.Disconnect("StartConnectionError")
	server.Handler.Clear()

	if _, ok := err.(*whatsapp.UnAuthorizedError); ok {
		server.Log.Warningf("unauthorized, setting unverified")
		return server.MarkVerified(false)
	}

	server.Log.Errorf("error on starting whatsapp server connection: %s", err.Error())
	return err
}

func (server *QpWhatsappServer) Stop(cause string) (err error) {
	if !server.stopRequested {

		// setting token
		server.stopRequested = true

		// loggging properly
		server.Log.Infof("stopping server: %s", cause)

		server.Disconnect("stop: " + cause)

		if server.Handler != nil {
			server.Handler.Clear()
		}
	}

	return
}

func (server *QpWhatsappServer) Restart() (err error) {
	err = server.Stop("restart")
	if err != nil {
		return
	}

	// wait 1 second before continue
	time.Sleep(1 * time.Second)

	server.Log.Info("re-initializing whatsapp server ...")
	return server.Start()
}

// Somente usar em caso de não ser permitida a reconxão automática
func (server *QpWhatsappServer) Disconnect(cause string) {
	if server.connection != nil && !server.connection.IsInterfaceNil() {
		if server.connection.IsConnected() {
			server.Log.Infof("disconnecting whatsapp server by: %s", cause)
			server.connection.Dispose(cause)
			server.connection = nil
		}
	}
}

// Retorna o titulo em cache (se houver) do id passado em parametro
func (server *QpWhatsappServer) GetChatTitle(wid string) string {
	return server.connection.GetChatTitle(wid)
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
	if status == whatsapp.Disconnected {
		return true
	}
	if status == whatsapp.Failed {
		return true
	}
	return false
}

func (server *QpWhatsappServer) GetWorking() bool {
	status := server.GetStatus()
	if status <= whatsapp.Stopped {
		return false
	} else if status == whatsapp.Disconnected {
		return false
	}
	return true
}

func (server *QpWhatsappServer) GetStatusString() string {
	return server.GetStatus().String()
}

func (server *QpWhatsappServer) ID() string {
	return server.WId
}

// Traduz o Wid para um número de telefone em formato E164
func (server *QpWhatsappServer) GetNumber() string {
	return library.GetPhoneByWId(server.WId)
}

func (server *QpWhatsappServer) GetTimestamp() time.Time {
	return server.Timestamp
}

func (server *QpWhatsappServer) GetStartedTime() time.Time {
	return server.StartTime
}

func (server *QpWhatsappServer) GetBatteryInfo() WhatsAppBateryStatus {
	if server.Battery != nil {
		return *server.Battery
	} else {
		return WhatsAppBateryStatus{}
	}
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
	return ENV.IsDevelopment()
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
func (server *QpWhatsappServer) UpdateToken(value string) (err error) {
	if len(value) == 0 {
		err = fmt.Errorf("empty token")
		return
	}

	err = server.UpdateToken(value)
	if err != nil {
		return
	}

	server.Log.Infof("updating token: %v", value)
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
func (server *QpWhatsappServer) Save() (err error) {
	server.Log.Infof("saving server info: %v", server)
	ok, err := server.db.Exists(server.Token)
	if err != nil {
		log.Errorf("error on checking existent server: %s", err.Error())
		return
	}

	// updating timestamp
	server.Timestamp = time.Now().UTC()

	if ok {
		server.Log.Debugf("updating server info: %v", server)
		return server.db.Update(server.QpServer)
	} else {
		server.Log.Debugf("adding server info: %v", server)
		return server.db.Add(server.QpServer)
	}
}

func (server *QpWhatsappServer) MarkVerified(value bool) (err error) {
	if server.Verified != value {
		server.Verified = value
		return server.Save()
	}
	return nil
}

func (server *QpWhatsappServer) ToggleGroups() (handle bool, err error) {
	server.HandleGroups = !server.HandleGroups
	return server.HandleGroups, server.Save()
}

func (server *QpWhatsappServer) ToggleBroadcast() (handle bool, err error) {
	server.HandleBroadcast = !server.HandleBroadcast
	return server.HandleBroadcast, server.Save()
}

func (server *QpWhatsappServer) ToggleDevel() (handle bool, err error) {
	server.Devel = !server.Devel

	if server.Devel {
		server.Log.Logger.SetLevel(log.DebugLevel)
	} else {
		server.Log.Logger.SetLevel(log.InfoLevel)
	}

	return server.Devel, server.Save()
}

//endregion

func (server *QpWhatsappServer) Delete() (err error) {
	if server.connection != nil {
		err = server.connection.Delete()
		if err != nil {
			return
		}

		server.connection = nil
	}

	return server.db.Delete(server.Token)
}

//endregion
//#region SEND

// Default send message method
func (server *QpWhatsappServer) SendMessage(msg *whatsapp.WhatsappMessage) (response whatsapp.IWhatsappSendResponse, err error) {
	server.Log.Debugf("sending msg to: %s", msg.Chat.Id)

	// leading with wrongs digit 9
	if ENV.ShouldRemoveDigit9() {
		msg.Chat.Id = library.RemoveDigit9(msg.Chat.Id)
	}

	if msg.HasAttachment() {
		if len(msg.Text) > 0 {

			// Overriding filename with caption text if IMAGE or VIDEO
			if msg.Type == whatsapp.ImageMessageType || msg.Type == whatsapp.VideoMessageType {
				msg.Attachment.FileName = msg.Text
			} else {

				// Copying and send text before file
				textMsg := *msg
				textMsg.Type = whatsapp.TextMessageType
				textMsg.Attachment = nil
				response, err = server.connection.Send(&textMsg)
				if err != nil {
					return
				} else {
					server.Handler.Message(&textMsg)
				}
			}
		}
	}

	// sending default msg
	response, err = server.connection.Send(msg)
	if err == nil {
		server.Handler.Message(msg)
	}
	return
}

//#endregion
//#region PROFILE PICTURE

func (server *QpWhatsappServer) GetProfilePicture(wid string, knowingId string) (picture *whatsapp.WhatsappProfilePicture, err error) {
	server.Log.Debugf("getting info about profile picture for: %s, with id: %s", wid, knowingId)

	return server.connection.GetProfilePicture(wid, knowingId)
}

//#endregion
//#region GROUP INVITE LINK

func (server *QpWhatsappServer) GetInvite(groupId string) (link string, err error) {
	return server.connection.GetInvite(groupId)
}

//#endregion
