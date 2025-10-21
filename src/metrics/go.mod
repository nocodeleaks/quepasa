module github.com/nocodeleaks/quepasa/metrics

replace github.com/nocodeleaks/quepasa/environment => ../environment

replace github.com/nocodeleaks/quepasa/webserver => ../webserver

replace github.com/nocodeleaks/quepasa/library => ../library

replace github.com/nocodeleaks/quepasa/whatsapp => ../whatsapp

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/nocodeleaks/quepasa/environment v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/webserver v0.0.0-00010101000000-000000000000
	github.com/prometheus/client_golang v1.23.2
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nocodeleaks/quepasa/library v0.0.0-00010101000000-000000000000 // indirect
	github.com/nocodeleaks/quepasa/whatsapp v0.0.0-00010101000000-000000000000 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/sys v0.37.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

go 1.24.0

toolchain go1.24.2
