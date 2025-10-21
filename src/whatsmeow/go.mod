module github.com/nocodeleaks/quepasa/whatsmeow

replace github.com/nocodeleaks/quepasa/whatsmeow => ./

replace github.com/nocodeleaks/quepasa/whatsapp => ../whatsapp

replace github.com/nocodeleaks/quepasa/library => ../library

replace github.com/nocodeleaks/quepasa/metrics => ../metrics

replace github.com/nocodeleaks/quepasa/environment => ../environment

replace github.com/nocodeleaks/quepasa/webserver => ../webserver

go 1.24.0

toolchain go1.24.2

require (
	github.com/nocodeleaks/quepasa/library v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/metrics v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/whatsapp v0.0.0-00010101000000-000000000000
)

require (
	github.com/beeper/argo-go v1.1.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/elliotchance/orderedmap/v3 v3.1.0 // indirect
	github.com/go-chi/chi/v5 v5.2.3 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nocodeleaks/quepasa/environment v0.0.0-00010101000000-000000000000 // indirect
	github.com/nocodeleaks/quepasa/webserver v0.0.0-00010101000000-000000000000 // indirect
	github.com/petermattis/goid v0.0.0-20250904145737-900bdf8bb490 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/vektah/gqlparser/v2 v2.5.30 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/exp v0.0.0-20251009144603-d2f985daa21b // indirect
	golang.org/x/text v0.30.0 // indirect
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/gosimple/slug v1.13.1
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/rs/zerolog v1.34.0 // indirect
	github.com/sirupsen/logrus v1.9.3
	go.mau.fi/libsignal v0.2.1 // indirect
	go.mau.fi/util v0.9.2
	go.mau.fi/whatsmeow v0.0.0-20251016095441-02c50743e601
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/net v0.46.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	google.golang.org/protobuf v1.36.10
)
