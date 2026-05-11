package models

import "strings"

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
			if ResolveRabbitMQClient(config.ConnectionString) {
				logentry.Infof("RabbitMQ connection initialized successfully: %s", config.ConnectionString)
			} else {
				logentry.Warnf("failed to initialize RabbitMQ connection: %s", config.ConnectionString)
			}
		}
	}
}

//#endregion
