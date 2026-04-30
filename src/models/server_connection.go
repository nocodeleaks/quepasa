package models

import (
	"fmt"
	"time"

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// Initialize starts the server asynchronously; errors are only logged.
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

	// Keep outbound dispatching subscriber attached once per handler lifecycle.
	if source.Handler != nil && !source.Handler.HasDispatchingSubscriber() {
		source.Handler.Register(NewOutboundDispatchingSubscriber(source))
	}
}

func (source *QpWhatsappServer) EnsureUnderlying() (err error) {

	if len(source.GetWId()) > 0 && !source.Verified {
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
			Wid:             source.GetWId(),
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

	// reset any pending stop intent so the session can run normally
	source.Intent = SessionIntentNone

	// Update start timestamp
	source.Timestamps.Start = time.Now().UTC()

	if !source.Handler.HasDispatchingSubscriber() {

		// Ensure outbound dispatching is attached even when other subscribers exist.
		source.Handler.Register(NewOutboundDispatchingSubscriber(source))
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

	// reset any pending stop intent so the session can run normally
	source.Intent = SessionIntentNone

	if !source.Handler.HasDispatchingSubscriber() {
		logger.Info("attaching handlers")

		// Ensure outbound dispatching is attached even when other subscribers exist.
		source.Handler.Register(NewOutboundDispatchingSubscriber(source))
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

	source.DisposeConnection("StartConnectionError")
	source.Handler.Clear()

	if _, ok := err.(*whatsapp.UnAuthorizedError); ok {
		logger.Warningf("unauthorized, setting unverified")
		return source.MarkVerified(false)
	}

	logger.Errorf("error on starting whatsapp server connection: %s", err.Error())
	return err
}

func (source *QpWhatsappServer) Stop(cause string) (err error) {
	if !source.Intent.IsStopRequested() {

		// mark stop intent so the session is not stopped twice
		source.Intent = SessionIntentStop

		// loggging properly
		logentry := source.GetLogger()
		logentry.Infof("stopping server: %s", cause)

		// Send stop event to dispatchers before disconnecting
		if source.Handler != nil {
			source.Handler.OnStopped(cause)
		}

		source.DisposeConnection("stop: " + cause)

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
func (source *QpWhatsappServer) DisposeConnection(cause string) {
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
	return server.GetUser()
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
	return server.GetWId()
}

// Traduz o Wid para um número de telefone em formato E164
// Ex: 5521967609494
func (server *QpWhatsappServer) GetNumber() string {
	return library.GetPhoneByWId(server.GetWId())
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
