module github.com/nocodeleaks/quepasa/metrics

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/nocodeleaks/quepasa/environment v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/webserver v0.0.0-00010101000000-000000000000
	github.com/prometheus/client_golang v1.16.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/nocodeleaks/quepasa/library v0.0.0-00010101000000-000000000000 // indirect
	github.com/nocodeleaks/quepasa/whatsapp v0.0.0-00010101000000-000000000000 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	golang.org/x/sys v0.40.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/nocodeleaks/quepasa/api => ../api

replace github.com/nocodeleaks/quepasa/environment => ../environment

replace github.com/nocodeleaks/quepasa/form => ../form

replace github.com/nocodeleaks/quepasa/library => ../library

replace github.com/nocodeleaks/quepasa/media => ../media

replace github.com/nocodeleaks/quepasa/metrics => ../metrics

replace github.com/nocodeleaks/quepasa/models => ../models

replace github.com/nocodeleaks/quepasa/rabbitmq => ../rabbitmq

replace github.com/nocodeleaks/quepasa/signalr => ../signalr

replace github.com/nocodeleaks/quepasa/sipproxy => ../sipproxy

replace github.com/nocodeleaks/quepasa/swagger => ../swagger

replace github.com/nocodeleaks/quepasa/webserver => ../webserver

replace github.com/nocodeleaks/quepasa/whatsapp => ../whatsapp

replace github.com/nocodeleaks/quepasa/whatsmeow => ../whatsmeow

go 1.24.0

toolchain go1.24.2
