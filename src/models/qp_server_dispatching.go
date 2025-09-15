package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Dispatching model for database operations
type QpServerDispatching struct {
	Context string                     `db:"context" json:"context"`
	db      QpDataDispatchingInterface `json:"-"`

	*QpDispatching
}

func (source *QpServerDispatching) Find(context string, connectionString string) (*QpServerDispatching, error) {
	return source.db.Find(context, connectionString)
}

func (source *QpServerDispatching) FindAll(context string) ([]*QpServerDispatching, error) {
	return source.db.FindAll(context)
}

func (source *QpServerDispatching) All() ([]*QpServerDispatching, error) {
	return source.db.All()
}

// passing extra info as json valid or default string
func (source *QpServerDispatching) GetExtraText() string {
	extraJson, err := json.Marshal(&source.Extra)
	if err != nil {
		return fmt.Sprintf("%v", source.Extra)
	} else {
		return string(extraJson)
	}
}

// trying to get interface from json string or default string
func (source *QpServerDispatching) ParseExtra() {
	extraText := fmt.Sprintf("%v", source.Extra)

	var extraJson interface{}
	err := json.Unmarshal([]byte(extraText), &extraJson)
	if err != nil {
		source.Extra = extraText
	} else {
		source.Extra = extraJson
	}
}

func (source *QpServerDispatching) Add(element *QpServerDispatching) error {
	return source.db.Add(element)
}

func (source *QpServerDispatching) Remove(context string, connectionString string) error {
	return source.db.Remove(context, connectionString)
}

func (source *QpServerDispatching) Clear(context string) error {
	return source.db.Clear(context)
}

// Implement QpDispatchingInterface

func (source *QpServerDispatching) GetConnectionString() string {
	return source.ConnectionString
}

func (source *QpServerDispatching) GetFailure() *time.Time {
	return source.Failure
}
