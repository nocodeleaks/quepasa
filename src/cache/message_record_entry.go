package cache

type MessageRecordEntry struct {
	Key    string        `json:"key"`
	Record MessageRecord `json:"record"`
}