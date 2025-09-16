package models

type QpDataDispatchingInterface interface {
	Find(context string, connectionString string) (*QpServerDispatching, error)
	FindAll(context string) ([]*QpServerDispatching, error)
	All() ([]*QpServerDispatching, error)
	Add(element *QpServerDispatching) error
	Update(element *QpServerDispatching) error
	UpdateContext(element *QpServerDispatching, context string) error
	Remove(context string, connectionString string) error
	Clear(context string) error

	// Unified dispatching methods for both webhook and RabbitMQ
	DispatchingAddOrUpdate(context string, dispatching *QpDispatching) (affected uint, err error)
	DispatchingRemove(context string, connectionString string) (affected uint, err error)
	DispatchingClear(context string) (err error)

	// Compatibility methods for interface - converts dispatching to legacy format
	GetWebhooks() []*QpWebhook
	GetRabbitMQConfigs() []*QpRabbitMQConfig
}
