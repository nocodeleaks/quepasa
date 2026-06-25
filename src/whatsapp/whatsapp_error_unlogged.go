package whatsapp

import (
	"fmt"
)

type UnLoggedError struct {
	Inner error
}

func (e *UnLoggedError) Error() string {
	return fmt.Sprintf("UnLogged: %s", e.Inner)
}
