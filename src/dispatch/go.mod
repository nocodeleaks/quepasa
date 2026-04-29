module github.com/nocodeleaks/quepasa/dispatch

replace github.com/nocodeleaks/quepasa/dispatch => ./

replace github.com/nocodeleaks/quepasa/rabbitmq => ../rabbitmq

replace github.com/nocodeleaks/quepasa/whatsapp => ../whatsapp

go 1.25.0

require (
	github.com/nocodeleaks/quepasa/rabbitmq v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/whatsapp v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
)
