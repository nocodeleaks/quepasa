module github.com/nocodeleaks/quepasa/whatsmeow

require (
	github.com/nocodeleaks/quepasa/library v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/whatsapp v0.0.0-00010101000000-000000000000
)

require (
	github.com/petermattis/goid v0.0.0-20250508124226-395b08cebbdb // indirect
	golang.org/x/exp v0.0.0-20250506013437-ce4c2cf36ca6 // indirect
	golang.org/x/text v0.25.0 // indirect
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/gosimple/slug v1.13.1
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v1.14.28
	github.com/rs/zerolog v1.34.0 // indirect
	github.com/sirupsen/logrus v1.9.3
	go.mau.fi/libsignal v0.2.0 // indirect
	go.mau.fi/util v0.8.7
	go.mau.fi/whatsmeow v0.0.0-20250521125706-91ac75c2f61a
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	google.golang.org/protobuf v1.36.6
)

replace github.com/nocodeleaks/quepasa/whatsmeow => ./

replace github.com/nocodeleaks/quepasa/whatsapp => ../whatsapp

replace github.com/nocodeleaks/quepasa/library => ../library

go 1.23.0

toolchain go1.23.2
