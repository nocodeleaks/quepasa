module github.com/nocodeleaks/quepasa/whatsmeow

require (
	github.com/nocodeleaks/quepasa/library v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/whatsapp v0.0.0-00010101000000-000000000000
)

require (
	github.com/petermattis/goid v0.0.0-20250319124200-ccd6737f222a // indirect
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0 // indirect
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/gosimple/slug v1.13.1
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/rs/zerolog v1.34.0 // indirect
	github.com/sirupsen/logrus v1.9.3
	go.mau.fi/libsignal v0.1.2 // indirect
	go.mau.fi/util v0.8.6 // indirect
	go.mau.fi/whatsmeow v0.0.0-20250501130609-4c93ee4e6efa
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/protobuf v1.36.6
)

replace github.com/nocodeleaks/quepasa/whatsmeow => ./

replace github.com/nocodeleaks/quepasa/whatsapp => ../whatsapp

replace github.com/nocodeleaks/quepasa/library => ../library

go 1.23.0

toolchain go1.23.2
