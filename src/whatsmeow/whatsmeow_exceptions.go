package whatsmeow

type WhatsmeowStoreNotFoundException struct {
	Wid string
}

func (e *WhatsmeowStoreNotFoundException) Error() string {
	return "cant find a store"
}

func (e *WhatsmeowStoreNotFoundException) Unauthorized() bool {
	return true
}
