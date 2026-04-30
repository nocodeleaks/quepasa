package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// RealtimePresenceChecker abstracts realtime connection lookup away from models.
type RealtimePresenceChecker interface {
	HasActiveConnections(token string) bool
}

// GlobalRealtimePresenceChecker can be wired at startup by transport modules.
var GlobalRealtimePresenceChecker RealtimePresenceChecker

// GlobalRabbitMQClientResolver allows transport-layer wiring without importing
// rabbitmq directly in the models package.
var GlobalRabbitMQClientResolver = func(connectionString string) bool {
	return false
}

type QpWhatsappServer struct {
	*QpServer
	QpDataDispatching // new dispatching system

	// should auto reconnect, false for qrcode scanner
	Reconnect bool `json:"reconnect"`

	connection     whatsapp.IWhatsappConnection `json:"-"`
	syncConnection *sync.Mutex                  `json:"-"` // Objeto de sinaleiro para evitar chamadas simultâneas a este objeto
	syncMessages   *sync.Mutex                  `json:"-"` // Objeto de sinaleiro para evitar chamadas simultâneas a este objeto

	Timestamps QpTimestamps `json:"timestamps"`

	Handler        *DispatchingHandler `json:"-"`
	GroupManager   *QpGroupManager     `json:"-"` // composition for group operations
	StatusManager  *QpStatusManager    `json:"-"` // composition for status operations
	ContactManager *QpContactManager   `json:"-"` // composition for contact operations

	// Intent tracks the current application-level lifecycle request for this session.
	// Use IsStopRequested() / IsDeleteRequested() instead of reading the field directly.
	Intent SessionIntent          `json:"-"`
	db     QpDataServersInterface `json:"-"`
}

// MarshalJSON customizes JSON serialization to include only dispatching field instead of webhooks
func (source QpWhatsappServer) MarshalJSON() ([]byte, error) {
	// Create a custom struct to control serialization
	type CustomServer struct {
		whatsapp.WhatsappOptions
		Token       string           `json:"token"`
		Wid         string           `json:"wid,omitempty"`
		Verified    bool             `json:"verified"`
		Devel       bool             `json:"devel"`
		Metadata    QpMetadata       `json:"metadata,omitempty"`
		User        string           `json:"user,omitempty"`
		Timestamp   time.Time        `json:"timestamp,omitempty"`
		Reconnect   bool             `json:"reconnect"`
		StartTime   time.Time        `json:"starttime,omitempty"`
		Timestamps  QpTimestamps     `json:"timestamps"`
		Dispatching []*QpDispatching `json:"dispatching,omitempty"`
		Uptime      library.Duration `json:"uptime"`
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

	// Calculate uptime
	uptime := time.Duration(0)
	if !timestamps.Start.IsZero() {
		uptime = time.Since(timestamps.Start)
	}

	custom := CustomServer{
		WhatsappOptions: source.WhatsappOptions,
		Token:           source.Token,
		Wid:             source.GetWId(),
		Verified:        source.Verified,
		Devel:           source.Devel,
		Metadata:        source.Metadata,
		User:            source.GetUser(),
		Timestamp:       source.Timestamp,
		Reconnect:       source.Reconnect,
		StartTime:       timestamps.Start,
		Timestamps:      timestamps,
		Dispatching:     dispatchingData,
		Uptime:          library.Duration(uptime),
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
			server: server,
		}

		logentry := server.GetLogger()
		logentry.Debug("ensuring messages handler for server")

		// logging
		handler.LogEntry = logentry

		// Inject cache backend from centralized cache service
		InjectCacheBackendIntoHandler(handler)

		// updating
		server.Handler = handler
	}
}

func (server *QpWhatsappServer) HasSignalRActiveConnections() bool {
	if server == nil {
		return false // invalid state
	}

	if GlobalRealtimePresenceChecker == nil {
		return false
	}

	return GlobalRealtimePresenceChecker.HasActiveConnections(server.Token)
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

	if server.Intent.IsDeleteRequested() {
		return whatsapp.Stopping
	}

	if server.connection == nil {
		if server.Verified {
			if server.Intent.IsStopRequested() {
				return whatsapp.Stopped
			}
			return whatsapp.UnPrepared
		}

		return whatsapp.UnVerified
	} else {
		if server.Intent.IsStopRequested() {
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
// Ex: 5521967609494
func (server QpWhatsappServer) GetWId() string {
	return server.QpServer.GetWId()
}

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

			// Resolver call initializes transport connection pool when available.
			if GlobalRabbitMQClientResolver(config.ConnectionString) {
				logentry.Infof("RabbitMQ connection initialized successfully: %s", config.ConnectionString)
			} else {
				logentry.Warnf("failed to initialize RabbitMQ connection: %s", config.ConnectionString)
			}
		}
	}
}

//#endregion
