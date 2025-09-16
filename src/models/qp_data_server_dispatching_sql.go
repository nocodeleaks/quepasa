package models

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type QpDataServerDispatchingSql struct {
	db *sqlx.DB
}

func (source QpDataServerDispatchingSql) Find(context string, connectionString string) (response *QpServerDispatching, err error) {
	var result []QpServerDispatching
	err = source.db.Select(&result, "SELECT * FROM dispatching WHERE context = ? AND connection_string = ?", context, connectionString)
	if err != nil {
		return
	}

	// adjust extra information
	for _, element := range result {
		element.ParseExtra()
		response = &element
		break
	}

	return
}

func (source QpDataServerDispatchingSql) FindAll(context string) ([]*QpServerDispatching, error) {
	result := []*QpServerDispatching{}
	err := source.db.Select(&result, "SELECT * FROM dispatching WHERE context = ?", context)

	// adjust extra information
	for _, element := range result {
		element.ParseExtra()
	}
	return result, err
}

func (source QpDataServerDispatchingSql) All() ([]*QpServerDispatching, error) {
	result := []*QpServerDispatching{}
	err := source.db.Select(&result, "SELECT * FROM dispatching")

	// adjust extra information
	for _, element := range result {
		element.ParseExtra()
	}
	return result, err
}

func (source QpDataServerDispatchingSql) Add(element *QpServerDispatching) error {
	query := `INSERT OR IGNORE INTO dispatching (context, connection_string, type, forwardinternal, trackid, readreceipts, groups, broadcasts, extra) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := source.db.Exec(query, element.Context, element.ConnectionString, element.Type, element.ForwardInternal, element.TrackId, element.ReadReceipts, element.Groups, element.Broadcasts, element.GetExtraText())
	return err
}

func (source QpDataServerDispatchingSql) Update(element *QpServerDispatching) error {
	query := `UPDATE dispatching SET type = ?, forwardinternal = ?, trackid = ?, readreceipts = ?, groups = ?, broadcasts = ?, extra = ? WHERE context = ? AND connection_string = ?`
	_, err := source.db.Exec(query, element.Type, element.ForwardInternal, element.TrackId, element.ReadReceipts, element.Groups, element.Broadcasts, element.GetExtraText(), element.Context, element.ConnectionString)
	return err
}

func (source QpDataServerDispatchingSql) UpdateContext(element *QpServerDispatching, context string) error {
	query := `UPDATE dispatching SET context = ? WHERE context = ? AND connection_string = ?`
	_, err := source.db.Exec(query, context, element.Context, element.ConnectionString)
	if err != nil {
		element.Context = context
	}
	return err
}

func (source QpDataServerDispatchingSql) Remove(context string, connectionString string) error {
	query := `DELETE FROM dispatching WHERE context = ? AND connection_string = ?`
	_, err := source.db.Exec(query, context, connectionString)
	return err
}

func (source QpDataServerDispatchingSql) RemoveWithResult(context string, connectionString string) (sql.Result, error) {
	query := `DELETE FROM dispatching WHERE context = ? AND connection_string = ?`
	return source.db.Exec(query, context, connectionString)
}

func (source QpDataServerDispatchingSql) Clear(context string) error {
	query := `DELETE FROM dispatching WHERE context = ?`
	_, err := source.db.Exec(query, context)
	return err
}

// GetWebhooks converts webhook dispatchings to QpWebhook format for interface compatibility
func (source QpDataServerDispatchingSql) GetWebhooks() []*QpWebhook {
	result := []*QpServerDispatching{}
	err := source.db.Select(&result, "SELECT * FROM dispatching WHERE type = 'webhook'")
	if err != nil {
		return []*QpWebhook{}
	}

	webhooks := []*QpWebhook{}
	for _, dispatching := range result {
		dispatching.ParseExtra()
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
			Wid:             dispatching.Context, // Context é o token do servidor
		}

		webhooks = append(webhooks, webhook)
	}

	return webhooks
}

// GetRabbitMQConfigs retorna apenas os registros do tipo rabbitmq
func (source QpDataServerDispatchingSql) GetRabbitMQConfigs() []*QpRabbitMQConfig {
	result := []*QpServerDispatching{}
	err := source.db.Select(&result, "SELECT * FROM dispatching WHERE type = 'rabbitmq'")
	if err != nil {
		return []*QpRabbitMQConfig{}
	}

	configs := []*QpRabbitMQConfig{}
	for _, dispatching := range result {
		dispatching.ParseExtra()
		config := &QpRabbitMQConfig{
			ConnectionString: dispatching.ConnectionString,
			ExchangeName:     "quepasa.exchange", // Fixo para todos
			RoutingKey:       "fixed",            // Fixo para todos
			TrackId:          dispatching.TrackId,
			Extra:            dispatching.Extra,
			Wid:              dispatching.Context,
		}

		// Copiar WhatsappOptions
		config.ReadReceipts = dispatching.ReadReceipts
		config.Groups = dispatching.Groups
		config.Broadcasts = dispatching.Broadcasts

		configs = append(configs, config)
	}

	return configs
}

// DispatchingAddOrUpdate adiciona ou atualiza um dispatching
func (source QpDataServerDispatchingSql) DispatchingAddOrUpdate(context string, dispatching *QpDispatching) (affected uint, err error) {
	if dispatching == nil || len(dispatching.ConnectionString) == 0 {
		err = fmt.Errorf("empty or nil dispatching")
		return
	}

	// Converter para QpServerDispatching
	serverDispatching := &QpServerDispatching{
		Context:       context,
		QpDispatching: dispatching,
	}

	// Verificar se já existe
	existing, err := source.Find(context, dispatching.ConnectionString)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return
	}

	if existing != nil {
		// Atualizar existente
		err = source.Update(serverDispatching)
		if err == nil {
			affected = 1
		}
	} else {
		// Adicionar novo
		err = source.Add(serverDispatching)
		if err == nil {
			affected = 1
		}
	}

	return
}

// DispatchingRemove remove um dispatching pelo connection_string e context
func (source QpDataServerDispatchingSql) DispatchingRemove(context string, connectionString string) (affected uint, err error) {
	if len(connectionString) == 0 {
		err = fmt.Errorf("empty connection string")
		return
	}

	result, err := source.RemoveWithResult(context, connectionString)
	if err != nil {
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return
	}

	affected = uint(rowsAffected)
	return
}

// DispatchingClear remove todos os dispatching de um contexto
func (source QpDataServerDispatchingSql) DispatchingClear(context string) (err error) {
	return source.Clear(context)
}
