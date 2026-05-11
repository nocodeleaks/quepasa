package cache

type MessagesBackend interface {
	Get(key string) (MessageRecord, bool, error)
	Set(key string, record MessageRecord) error
	Delete(key string) error
	List() ([]MessageRecordEntry, error)
	Close() error
}
