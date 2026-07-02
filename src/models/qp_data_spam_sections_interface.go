package models

type QpDataSpamSectionsInterface interface {
	Find(token string) (*QpSpamSection, error)
	ListAll() ([]*QpSpamSection, error)
	Upsert(section *QpSpamSection) error
	UpdatePosition(token string, position int) error
	Delete(token string) (bool, error)
	NextPosition() (int, error)
}
