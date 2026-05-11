package models

type QpDataConversationLabelsInterface interface {
	FindAllForUser(user string, activeOnly *bool) ([]*QpConversationLabel, error)
	FindByIDForUser(id int64, user string) (*QpConversationLabel, error)
	Create(label *QpConversationLabel) (*QpConversationLabel, error)
	Update(label *QpConversationLabel) error
	Delete(id int64, user string) error
	Assign(serverToken string, chatID string, labelID int64, user string) (uint, error)
	Remove(serverToken string, chatID string, labelID int64, user string) (uint, error)
	FindConversationLabels(serverToken string, chatID string, user string) ([]*QpConversationLabel, error)
	FindConversationLabelsMap(serverToken string, user string, chatIDs []string) (map[string][]*QpConversationLabel, error)
}
