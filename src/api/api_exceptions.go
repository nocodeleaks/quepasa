package api

import (
	"fmt"
	"strings"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type ApiException interface {
	error
	Prepend(string)
}

type ApiExceptionBase struct {
	Inner    error
	Messages []string
}

func prepend(x []string, y string) []string {
	x = append(x, "")
	copy(x[1:], x)
	x[0] = y
	return x
}

func (e *ApiExceptionBase) Prepend(message string) {
	e.Messages = prepend(e.Messages, message)
}

func (e *ApiExceptionBase) Prependf(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	e.Prepend(message)
}

func (e *ApiExceptionBase) Error() string {
	var message string
	message += strings.Join(e.Messages[:], ", ")
	if e.Inner != nil {
		if len(message) > 0 {
			message += ": "
		}
		message += e.Inner.Error()
	}
	return fmt.Sprint(message)
}

type ApiServerNotReadyException struct {
	Wid    string
	Status whatsapp.WhatsappConnectionState
}

func (e *ApiServerNotReadyException) Error() string {
	return fmt.Sprintf("server (%s) not ready yet ! current status: %s.", e.Wid, e.Status.String())
}

type BadRequestException struct {
	ApiExceptionBase
}
