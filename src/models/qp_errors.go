package models

import (
	"errors"
	"fmt"
)

var ErrInvalidConnection error = errors.New("nil or invalid connection state")
var ErrFormUnauthenticated error = errors.New("missing user id")

type ErrServiceUnreachable struct {
	Server  string
	Message string
}

func (e *ErrServiceUnreachable) Error() string {
	return fmt.Sprintf("(%s)(ERR) WhatsApp service is unreachable by '%s', probably an WhatsApp (Facebook) servers error", e.Server, e.Message)
}
