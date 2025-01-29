package whatsapp

import (
	"fmt"
)

type UnAuthorizedError struct {
	Inner error
}

func (e *UnAuthorizedError) Error() string {
	return fmt.Sprintf("UnAuthorized: %s", e.Inner)
}
