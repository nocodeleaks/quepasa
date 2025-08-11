package whatsmeow

import (
	"context"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	whatsmeow "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	types "go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type WhatsmeowServiceModel struct {
	Container *sqlstore.Container
	Options   WhatsmeowOptions

	library.LogStruct
}

// get default log entry, never nil
func (source *WhatsmeowServiceModel) GetLogger() *log.Entry {
	if source.LogEntry == nil {
		logentry := log.WithContext(context.Background())
		logentry.Level = log.ErrorLevel
		source.LogEntry = logentry
	}

	return source.LogEntry
}

var WhatsmeowService *WhatsmeowServiceModel

func Start(options WhatsmeowOptions, dbParameters library.DatabaseParameters, logentry *log.Entry) {
	logentry.Infof("starting Whatsmeow Service, with log level: %s", logentry.Level)
	if WhatsmeowService != nil {
		err := fmt.Errorf("whatsmeow service is already started, if you wanna change options, restart the service")
		panic(err)
	}

	dbloglevel := WhatsmeowDBLogLevel
	if len(options.DBLogLevel) > 0 {
		dbloglevel = options.DBLogLevel
	}
	dbLog := waLog.Stdout("whatsmeow/database", dbloglevel, true)

	if len(dbParameters.DataBase) == 0 {
		dbParameters.DataBase = "whatsmeow"
	}
	connectionString := dbParameters.GetConnectionString()
	container, err := sqlstore.New(context.TODO(), dbParameters.Driver, connectionString, dbLog)
	if err != nil {
		err = fmt.Errorf("error on creating db container: %s", err.Error())
		panic(err)
	}

	WhatsmeowService = &WhatsmeowServiceModel{
		Container: container,
		Options:   options,
	}

	// logging
	logentry = library.NewLogEntry(WhatsmeowService)
	if len(options.LogLevel) > 0 {
		loglevel, err := log.ParseLevel(options.LogLevel)
		if err == nil {
			logentry.Level = loglevel
		}
	} else {
		logentry.Level = WhatsmeowLogLevel
	}

	WhatsmeowService.LogEntry = logentry
	logentry.Infof("new Whatsmeow Service created, with log level: %s", logentry.Level)

	showing := whatsapp.WhatsappWebAppName

	// trim spaces from app name previous setted, if exists
	previousShowing := strings.TrimSpace(whatsapp.WhatsappWebAppSystem)
	if len(previousShowing) > 0 {
		// using previous setted name
		showing = previousShowing
	}

	var version [3]uint32
	version[0] = 0
	version[1] = 9
	version[2] = 0
	store.SetOSInfo(showing, version)

	// this section is broken, history sync configurations, do nothing ......
	// --------------------------------

	historysync := WhatsmeowService.GetHistorySync()
	if historysync != nil {
		HistorySyncValue := *historysync
		logentry.Infof("setting history sync to %v days", HistorySyncValue)

		if HistorySyncValue == 0 {
			store.DeviceProps.RequireFullSync = proto.Bool(true)
			store.DeviceProps.HistorySyncConfig = &waCompanionReg.DeviceProps_HistorySyncConfig{
				FullSyncDaysLimit:   proto.Uint32(3650),
				FullSyncSizeMbLimit: proto.Uint32(102400),
			}
		} else {
			store.DeviceProps.RequireFullSync = proto.Bool(false)
			store.DeviceProps.HistorySyncConfig = &waCompanionReg.DeviceProps_HistorySyncConfig{
				RecentSyncDaysLimit: proto.Uint32(HistorySyncValue * 10),
			}
		}

		store.DeviceProps.HistorySyncConfig.StorageQuotaMb = proto.Uint32(102400)
	}
}

func (source WhatsmeowServiceModel) GetServiceOptions() whatsapp.WhatsappOptionsExtended {
	return source.Options.WhatsappOptionsExtended
}

func (source *WhatsmeowServiceModel) GetHistorySync() *uint32 {
	return source.Options.HistorySync
}

// Used for scan QR Codes
// Dont forget to attach handlers after success login
func (source *WhatsmeowServiceModel) CreateEmptyConnection() (conn *WhatsmeowConnection, err error) {
	logentry := source.GetLogger()
	options := &whatsapp.WhatsappConnectionOptions{
		Reconnect: false,
		LogStruct: library.LogStruct{LogEntry: logentry},
	}
	return source.CreateConnection(options)
}

func (source *WhatsmeowServiceModel) CreateConnection(options *whatsapp.WhatsappConnectionOptions) (conn *WhatsmeowConnection, err error) {
	client, err := source.GetWhatsAppClient(options)
	if err != nil {
		return
	}

	logentry := options.GetLogger()
	client.EnableAutoReconnect = options.GetReconnect()

	if len(options.Wid) > 0 {
		logentry = options.LogWithField(LogFields.WId, options.Wid)
	}

	logentry.Infof("creating whatsmeow connection with log level: %s", logentry.Level)

	// Create connection first (will create handlers internally)
	conn = &WhatsmeowConnection{
		Client:    client,
		LogStruct: library.LogStruct{LogEntry: logentry},
	}

	// Initialize handlers with proper options after connection is created
	err = conn.initializeHandlers(options.WhatsappOptions, source.Options)
	if err != nil {
		return
	}

	client.PrePairCallback = conn.PairedCallBack

	// Configure AutoReconnectHook
	client.AutoReconnectHook = conn.AutoReconnectHook

	return
}

// Gets an existing store or create a new one for empty wid
func (service *WhatsmeowServiceModel) GetOrCreateStore(wid string) (str *store.Device, err error) {
	if wid == "" {
		str = service.Container.NewDevice()
	} else {
		devices, err := service.Container.GetAllDevices(context.TODO())
		if err != nil {
			err = fmt.Errorf("{Whatsmeow}{EX001} error on getting store from wid {%s}: %v", wid, err)
			return str, err
		} else {
			for _, element := range devices {
				if element.ID.String() == wid {
					str = element
					break
				}
			}

			if str == nil {
				err = &WhatsmeowStoreNotFoundException{Wid: wid}
				return str, err
			}
		}
	}

	return
}

// Temporary
func (service *WhatsmeowServiceModel) GetStoreForMigrated(phone string) (str *store.Device, err error) {

	devices, err := service.Container.GetAllDevices(context.TODO())
	if err != nil {
		err = fmt.Errorf("{Whatsmeow}{EX001} error on getting store from phone {%s}: %v", phone, err)
		return str, err
	} else {
		for _, element := range devices {
			if library.GetPhoneByWId(element.ID.String()) == phone {
				str = element
				break
			}
		}

		if str == nil {
			err = &WhatsmeowStoreNotFoundException{Wid: phone}
			return str, err
		}
	}

	return
}

func (source *WhatsmeowServiceModel) GetWhatsAppClient(options *whatsapp.WhatsappConnectionOptions) (client *whatsmeow.Client, err error) {
	loglevel := WhatsmeowClientLogLevel
	_, logerr := log.ParseLevel(source.Options.WMLogLevel)
	if logerr == nil {
		loglevel = source.Options.WMLogLevel
	}

	wid := options.GetWid()
	clientLog := waLog.Stdout("whatsmeow/client", loglevel, true)
	if len(wid) > 0 {
		clientLog = clientLog.Sub(wid)
	}

	deviceStore, err := source.GetOrCreateStore(wid)
	if deviceStore != nil {
		client = whatsmeow.NewClient(deviceStore, clientLog)
		client.AutoTrustIdentity = true
		client.EnableAutoReconnect = options.GetReconnect()
	}
	return
}

// Flush entire Whatsmeow Database
// Use with wisdom !
func (service *WhatsmeowServiceModel) FlushDatabase() (err error) {
	devices, err := service.Container.GetAllDevices(context.TODO())
	if err != nil {
		return
	}

	for _, element := range devices {
		err = element.Delete(context.TODO())
		if err != nil {
			return
		}
	}

	return
}

// GetPhoneFromLID returns the phone number for a given @lid
func (service *WhatsmeowServiceModel) GetPhoneFromLID(lid string) (string, error) {
	if service == nil || service.Container == nil {
		return "", fmt.Errorf("service or container not initialized")
	}

	// Parse the LID to JID format
	lidJID, err := types.ParseJID(lid)
	if err != nil {
		return "", fmt.Errorf("invalid LID format: %v", err)
	}

	// Get all devices to search through their LID mappings
	devices, err := service.Container.GetAllDevices(context.TODO())
	if err != nil {
		return "", fmt.Errorf("failed to get devices: %v", err)
	}

	// Search through all devices for the LID mapping
	for _, device := range devices {
		// Get the corresponding phone number from the device
		phoneJID, err := device.LIDs.GetPNForLID(context.TODO(), lidJID)
		if err == nil && !phoneJID.IsEmpty() {
			return phoneJID.User, nil
		}
	}

	return "", fmt.Errorf("no phone found for LID %s", lid)

}
