package models

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

// Serviço que controla os servidores / bots individuais do whatsapp
type QPWhatsappService struct {
	Servers     map[string]*QpWhatsappServer `json:"-"`
	DB          *QpDatabase                  `json:"-"`
	Initialized bool                         `json:"-"`

	initlock   *sync.Mutex `json:"-"`
	appendlock *sync.Mutex `json:"-"`

	library.LogStruct
}

var WhatsappService *QPWhatsappService

func QPWhatsappStart(logentry *log.Entry) error {
	if WhatsappService == nil {
		logentry.Infof("whatsapp service starting, with log level: %s", logentry.Level)

		db := GetDatabase()
		WhatsappService = &QPWhatsappService{
			Servers:    make(map[string]*QpWhatsappServer),
			DB:         db,
			initlock:   &sync.Mutex{},
			appendlock: &sync.Mutex{},
		}

		loglevel := logentry.Level
		logentry = library.NewLogEntry(WhatsappService)
		logentry.Level = loglevel
		WhatsappService.LogEntry = logentry

		// seeding database
		err := InitialSeed()
		if err != nil {
			return err
		}
		// iniciando servidores e cada bot individualmente
		return WhatsappService.Initialize()
	} else {
		logentry.Debug("attempt to start whatsapp service, already started ...")
	}
	return nil
}

// Inclui um novo servidor em um serviço já em andamento
// *Usado quando se passa pela verificação do QRCode
// *Usado quando se inicializa o sistema
func (source *QPWhatsappService) AppendNewServer(info *QpServer) (server *QpWhatsappServer, err error) {
	logentry := source.GetLogger()

	// checking if it is cached already
	server, ok := source.Servers[info.Token]
	if !ok {
		// adding to cache
		logentry.Infof("adding new server on cache: %s, wid: %s", info.Token, info.Wid)

		// Creating a new instance
		server, err = source.NewQpWhatsappServer(info)
		if err != nil {
			logentry.Errorf("error on append new server: %s, :: %s", info.Wid, err.Error())
			return
		}

		source.Servers[info.Token] = server
	} else {
		// updating cached item
		logentry.Infof("updating new server on cache: %s, wid: %s", info.Token, info.Wid)

		server.QpServer = info
	}
	return
}

func (source *QPWhatsappService) AppendPaired(paired *QpWhatsappPairing) (server *QpWhatsappServer, err error) {
	logger := source.GetLogger()

	// checking if it is cached already
	server, ok := source.Servers[paired.Token]
	if !ok {
		// adding to cache
		logger.Infof("adding paired server on cache: %s, wid: %s", paired.Token, paired.Wid)

		info := &QpServer{Token: paired.Token, Wid: paired.Wid}

		// Creating a new instance
		server, err = source.NewQpWhatsappServer(info)
		if err != nil {
			logger.Errorf("error on append new server: %s, :: %s", info.Wid, err.Error())
			return
		}

		source.Servers[info.Token] = server
	} else {
		server.Token = paired.Token
		server.Wid = paired.Wid

		// updating cached item
		logger.Infof("updating paired server on cache: %s, old wid: %s, new wid: %s", server.Token, server.Wid, paired.Wid)
	}

	server.connection = paired.conn
	server.Verified = true

	// Update handler on the existing connection
	// The connection was created during pairing without the server's handler
	// We must link the server's handler to receive messages properly
	if server.Handler != nil && server.connection != nil && !server.connection.IsInterfaceNil() {
		logger.Debug("updating server handler on paired connection")
		server.connection.UpdateHandler(server.Handler)
	}

	// checking user
	if len(paired.Username) > 0 {
		server.User = paired.Username
	}

	err = server.Save("server paired")
	return
}

//region CONSTRUCTORS

// Instance a new quepasa whatsapp server control
func (source *QPWhatsappService) NewQpWhatsappServer(info *QpServer) (server *QpWhatsappServer, err error) {

	logentry := source.GetLogger()
	logentry.Debug("creating a new QP Whatsapp Server")

	if info == nil {
		err = fmt.Errorf("missing server information")
		return
	}

	startTime := time.Now().UTC()
	server = &QpWhatsappServer{
		QpServer:       info,
		Reconnect:      true,
		syncConnection: &sync.Mutex{},
		syncMessages:   &sync.Mutex{},
		Timestamps: QpTimestamps{
			Start: startTime,
		},

		StopRequested: false, // setting initial state
		db:            source.DB.Servers,
	}

	var loglevel log.Level
	if info.Devel {
		loglevel = log.DebugLevel
		if logentry.Level > loglevel {
			loglevel = logentry.Level
		}
	} else {
		loglevel = log.InfoLevel
	}

	serverLogEntry := library.NewLogEntry(server)
	serverLogEntry = serverLogEntry.WithField(LogFields.Token, info.Token)

	if len(info.Wid) > 0 {
		serverLogEntry = serverLogEntry.WithField(LogFields.WId, info.Wid)
	}

	serverLogEntry.Level = loglevel
	server.LogEntry = serverLogEntry
	logentry.Infof("server created, log level: %s", serverLogEntry.Level)
	logentry.Trace("server created ...")

	server.HandlerEnsure()
	server.DispatchingEnsure()
	server.DispatchingFill(info, source.DB.Dispatching)
	return
}

func (source *QPWhatsappService) GetOrCreateServerFromToken(token string) (server *QpWhatsappServer, err error) {
	logger := source.GetLogger()
	logger.Debugf("locating server: %s", token)

	server, ok := source.Servers[token]
	if !ok {
		logger.Debugf("server: %s, not in cache, looking up database", token)
		exists, err := source.DB.Servers.Exists(token)
		if err != nil {
			err = fmt.Errorf("whatsapp service, get or create server from token, database exists error: %s", err.Error())
			return nil, err
		}

		var info *QpServer
		if exists {
			info, err = source.DB.Servers.FindByToken(token)
			if err != nil {
				err = fmt.Errorf("whatsapp service, get or create server from token, database find error: %s", err.Error())
				return nil, err
			}
			logger.Debugf("server: %s, found", token)
		} else {
			info = &QpServer{
				Token: token,
			}
		}

		server, err = source.AppendNewServer(info)
		return server, err
	}

	return
}

/*
<summary>

	Get or Create a server for scanned qrcode from forms with current user informations and a whatsapp section id
	* use same token if already exists

</summary>
*/
func (service *QPWhatsappService) GetOrCreateServer(user string, wid string) (result *QpWhatsappServer, err error) {
	log.Debugf("locating server with section id: %s", wid)

	phone := library.GetPhoneByWId(wid)
	log.Infof("wid to phone: %s", phone)

	var server *QpWhatsappServer
	servers := service.GetServersForUser(user)
	for _, item := range servers {
		if item.GetNumber() == phone {
			server = item
			server.Wid = wid
			break
		}
	}

	if server == nil {
		token := uuid.New().String()
		log.Infof("creating new server with token: %s", token)
		info := &QpServer{
			Token: token,
			User:  user,
			Wid:   wid,
		}

		server, err = service.AppendNewServer(info)
		if err != nil {
			err = fmt.Errorf("whatsapp service, get or create on append error: %s", err.Error())
			return
		}
	}

	result = server
	return
}

// delete whatsapp server and remove from cache
func (service *QPWhatsappService) Delete(server *QpWhatsappServer) (err error) {
	err = server.Delete()
	if err != nil {
		err = fmt.Errorf("whatsapp service, delete error: %s", err.Error())
		return
	}

	delete(service.Servers, server.Token)
	return
}

// method that will initiate all servers from database
func (source *QPWhatsappService) Initialize() (err error) {

	if !source.Initialized {

		servers := source.DB.Servers.FindAll()
		for _, info := range servers {

			// appending server to cache
			server, err := source.AppendNewServer(info)
			if err != nil {
				err = fmt.Errorf("whatsapp service, initialize error: %s", err.Error())
				return err
			}

			logentry := source.GetLogger()

			state := server.GetStatus()
			if state == whatsapp.UnPrepared || IsValidToStart(state) {

				// initialize individual server
				logentry.Debugf("starting whatsapp server ... on %s state", state)
				go server.Initialize()
			} else {
				logentry.Debugf("not auto starting cause state: %s", state)
			}
		}

		source.Initialized = true
	}

	return
}

// Função privada que irá iniciar todos os servidores apartir do banco de dados
func (service *QPWhatsappService) GetServersForUser(username string) (servers map[string]*QpWhatsappServer) {
	servers = make(map[string]*QpWhatsappServer)
	for _, server := range service.Servers {
		if server.GetOwnerID() == username {
			servers[strings.ToLower(server.Token)] = server
		}
	}
	return
}

// Case insensitive
func (service *QPWhatsappService) FindByToken(token string) (*QpWhatsappServer, error) {
	for _, server := range service.Servers {
		if strings.EqualFold(server.Token, token) {
			return server, nil
		}
	}

	err := fmt.Errorf("server not found for token: %s", token)
	return nil, err
}

func (source *QPWhatsappService) GetUser(username string, password string) (user *QpUser, err error) {
	logger := source.GetLogger()
	logger.Debugf("finding user: %s", username)
	return source.DB.Users.Check(username, password)
}

//region CONTROLLER - HEALTH

func (source *QPWhatsappService) GetHealth() (items []QpHealthResponseItem) {
	for _, server := range source.Servers {
		item := ToHealthReponseItem(server)
		items = append(items, item)
	}
	return items
}

func ToHealthReponseItem(server *QpWhatsappServer) QpHealthResponseItem {
	state := server.GetState()
	return QpHealthResponseItem{
		Token:     server.Token,
		Wid:       server.Wid,
		State:     state,
		StateCode: int(state),
	}
}

//endregion
