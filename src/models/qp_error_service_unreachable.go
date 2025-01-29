package models

import (
	"fmt"
)

type ServiceUnreachableError struct {
    Server string
	Message string
}

func (e *ServiceUnreachableError) Error() string {
    return fmt.Sprintf("(%s)(ERR) WhatsApp service is unreachable by '%s', probably an WhatsApp (Facebook) servers error", e.Server, e.Message)
}