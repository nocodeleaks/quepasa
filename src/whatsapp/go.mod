module github.com/nocodeleaks/quepasa/whatsapp

require github.com/nocodeleaks/quepasa/library v0.0.0-00010101000000-000000000000

require github.com/sirupsen/logrus v1.9.3

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	go.mau.fi/libsignal v0.2.0 // indirect
	go.mau.fi/util v0.8.8 // indirect
	go.mau.fi/whatsmeow v0.0.0-20250709212552-0b8557ee0860 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/nocodeleaks/quepasa/whatsapp => ./

replace github.com/nocodeleaks/quepasa/library => ../library

go 1.23.0

toolchain go1.24.2
