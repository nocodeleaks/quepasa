module github.com/nocodeleaks/quepasa/signalr

replace github.com/nocodeleaks/quepasa/api => ../api

replace github.com/nocodeleaks/quepasa/apps/form => ../apps/form

replace github.com/nocodeleaks/quepasa/media => ../media

replace github.com/nocodeleaks/quepasa/metrics => ../metrics

replace github.com/nocodeleaks/quepasa/rabbitmq => ../rabbitmq

replace github.com/nocodeleaks/quepasa/signalr => ../signalr

replace github.com/nocodeleaks/quepasa/sipproxy => ../sipproxy

replace github.com/nocodeleaks/quepasa/swagger => ../swagger

replace github.com/nocodeleaks/quepasa/whatsmeow => ../whatsmeow

go 1.25.0

// Local module replacements
replace github.com/nocodeleaks/quepasa/environment => ../environment

replace github.com/nocodeleaks/quepasa/library => ../library

replace github.com/nocodeleaks/quepasa/models => ../models

replace github.com/nocodeleaks/quepasa/webserver => ../webserver

replace github.com/nocodeleaks/quepasa/whatsapp => ../whatsapp

require (
	// External dependencies
	github.com/go-chi/chi/v5 v5.2.3
	github.com/go-kit/log v0.2.1

	// Local dependencies
	github.com/nocodeleaks/quepasa/environment v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/webserver v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/whatsapp v0.0.0-00010101000000-000000000000
	github.com/philippseith/signalr v0.6.3
	github.com/sirupsen/logrus v1.9.3
)

require (
	// Indirect dependencies
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/nocodeleaks/quepasa/library v0.0.0-00010101000000-000000000000 // indirect
	github.com/teivah/onecontext v1.3.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	nhooyr.io/websocket v1.8.11 // indirect
)

require (
	github.com/google/uuid v1.6.0 // indirect
	go.uber.org/goleak v1.3.0 // indirect
	golang.org/x/net v0.50.0 // indirect
)
