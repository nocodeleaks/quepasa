package models

import (
	rabbitmq "github.com/nocodeleaks/quepasa/rabbitmq"
)

// Dispatching model
type QpDataDispatching struct {
	Dispatching []*QpDispatching `json:"dispatching,omitempty"`
	context     string
	db          QpDataDispatchingInterface
}

// Fill start memory cache
func (source *QpDataDispatching) DispatchingFill(info *QpServer, db QpDataDispatchingInterface) (err error) {
	dispatching := []*QpDispatching{}
	source.context = info.Token
	source.db = db

	logentry := info.GetLogger()
	logentry.Trace("getting dispatching from database")

	dispatchings, err := source.db.FindAll(source.context)
	if err != nil {
		logentry.Errorf("cannot find dispatching: %s", err.Error())
		return
	}

	for _, element := range dispatchings {

		// logging for running dispatching
		dispatchingLogEntry := logentry.WithField(LogFields.Url, element.ConnectionString)
		element.LogEntry = dispatchingLogEntry

		element.Wid = info.Wid
		dispatching = append(dispatching, element.QpDispatching)
	}

	logentry.Debugf("%v dispatching(s) found", len(dispatching))

	// updating
	source.Dispatching = dispatching
	return
}

func (source *QpDataDispatching) DispatchingAddOrUpdate(dispatching *QpDispatching) (affected uint, err error) {
	affected, err = source.db.DispatchingAddOrUpdate(source.context, dispatching)
	if err != nil {
		return
	}

	// Update memory cache
	exists := false
	for index, element := range source.Dispatching {
		if element.ConnectionString == dispatching.ConnectionString {
			source.Dispatching[index] = dispatching
			exists = true
			break
		}
	}

	if !exists {
		source.Dispatching = append(source.Dispatching, dispatching)
	}

	return
}

func (source *QpDataDispatching) DispatchingRemove(connectionString string) (affected uint, err error) {
	affected, err = source.db.DispatchingRemove(source.context, connectionString)
	if err != nil {
		return
	}

	// Close RabbitMQ client for this connection string before removing from memory
	var isRabbitMQ bool
	for _, element := range source.Dispatching {
		if element.ConnectionString == connectionString && element.IsRabbitMQ() {
			isRabbitMQ = true
			break
		}
	}

	// Update memory cache
	i := 0
	for _, element := range source.Dispatching {
		if element.ConnectionString != connectionString {
			source.Dispatching[i] = element
			i++
		}
	}
	source.Dispatching = source.Dispatching[:i]

	// Close the RabbitMQ client after removing from memory to avoid race conditions
	if isRabbitMQ {
		rabbitmq.CloseRabbitMQClient(connectionString)
	}

	return
}

func (source *QpDataDispatching) DispatchingClear() (err error) {
	// Close all RabbitMQ clients before clearing
	for _, element := range source.Dispatching {
		if element.IsRabbitMQ() {
			rabbitmq.CloseRabbitMQClient(element.ConnectionString)
		}
	}

	// Clear from database
	err = source.db.DispatchingClear(source.context)
	if err != nil {
		return
	}

	// Clear memory cache
	source.Dispatching = source.Dispatching[:0]

	return
}

// Get webhooks only (for backward compatibility)
func (source *QpDataDispatching) GetWebhooks() []*QpWebhook {
	webhooks := []*QpWebhook{}
	for _, dispatching := range source.Dispatching {
		if dispatching.IsWebhook() {
			webhook := &QpWebhook{
				LogStruct:       dispatching.LogStruct,
				WhatsappOptions: dispatching.WhatsappOptions,
				Url:             dispatching.ConnectionString,
				ForwardInternal: dispatching.ForwardInternal,
				TrackId:         dispatching.TrackId,
				Extra:           dispatching.Extra,
				Failure:         dispatching.Failure,
				Success:         dispatching.Success,
				Timestamp:       dispatching.Timestamp,
				Wid:             dispatching.Wid,
			}
			webhooks = append(webhooks, webhook)
		}
	}
	return webhooks
}

// Get RabbitMQ configs only
func (source *QpDataDispatching) GetRabbitMQConfigs() []*QpRabbitMQConfig {
	configs := []*QpRabbitMQConfig{}
	for _, dispatching := range source.Dispatching {
		if dispatching.IsRabbitMQ() {
			config := &QpRabbitMQConfig{
				LogStruct:        dispatching.LogStruct,
				WhatsappOptions:  dispatching.WhatsappOptions,
				ConnectionString: dispatching.ConnectionString,
				ForwardInternal:  dispatching.ForwardInternal,
				TrackId:          dispatching.TrackId,
				Extra:            dispatching.Extra,
				Failure:          dispatching.Failure,
				Success:          dispatching.Success,
				Timestamp:        dispatching.Timestamp,
				Wid:              dispatching.Wid,
			}
			configs = append(configs, config)
		}
	}
	return configs
}
