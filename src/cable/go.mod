module github.com/nocodeleaks/quepasa/cable

replace github.com/nocodeleaks/quepasa/media => ../media

replace github.com/nocodeleaks/quepasa/library => ../library

replace github.com/nocodeleaks/quepasa/metrics => ../metrics

replace github.com/nocodeleaks/quepasa/models => ../models

replace github.com/nocodeleaks/quepasa/whatsapp => ../whatsapp

replace github.com/nocodeleaks/quepasa/whatsmeow => ../whatsmeow

replace github.com/nocodeleaks/quepasa/rabbitmq => ../rabbitmq

replace github.com/nocodeleaks/quepasa/environment => ../environment

replace github.com/nocodeleaks/quepasa/api => ../api

replace github.com/nocodeleaks/quepasa/form => ../form

replace github.com/nocodeleaks/quepasa/signalr => ../signalr

replace github.com/nocodeleaks/quepasa/sipproxy => ../sipproxy

replace github.com/nocodeleaks/quepasa/swagger => ../swagger

replace github.com/nocodeleaks/quepasa/webserver => ../webserver

replace github.com/nocodeleaks/quepasa/cable => ../cable

go 1.25.0

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/go-chi/jwtauth v4.0.4+incompatible
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	github.com/nocodeleaks/quepasa/models v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/webserver v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/whatsapp v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
)
