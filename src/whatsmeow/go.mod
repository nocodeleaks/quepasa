module github.com/nocodeleaks/quepasa/whatsmeow

require (
	github.com/nocodeleaks/quepasa/library v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/whatsapp v0.0.0-00010101000000-000000000000
)

require (
	go.mau.fi/util v0.1.0 // indirect
	go.mau.fi/whatsmeow v0.0.0-20230916142552-a743fdc23bf1
)

require (
	filippo.io/edwards25519 v1.0.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/gosimple/slug v1.13.1
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/mattn/go-sqlite3 v1.14.17
	github.com/sirupsen/logrus v1.9.3
	go.mau.fi/libsignal v0.1.0 // indirect
	golang.org/x/crypto v0.13.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/protobuf v1.31.0
)

replace github.com/nocodeleaks/quepasa/whatsmeow => ./
replace github.com/nocodeleaks/quepasa/whatsapp => ../whatsapp
replace github.com/nocodeleaks/quepasa/library => ../library

go 1.20
