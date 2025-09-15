module github.com/nocodeleaks/quepasa/models

replace github.com/nocodeleaks/quepasa/media => ../media

replace github.com/nocodeleaks/quepasa/library => ../library

replace github.com/nocodeleaks/quepasa/whatsmeow => ../whatsmeow

replace github.com/nocodeleaks/quepasa/whatsapp => ../whatsapp

replace github.com/nocodeleaks/quepasa/rabbitmq => ../rabbitmq

replace github.com/nocodeleaks/quepasa/models => ./

replace github.com/nocodeleaks/quepasa/environment => ../environment

require (
	github.com/google/uuid v1.6.0
	github.com/jmoiron/sqlx v1.3.5
	github.com/joncalhoun/migrate v0.0.2
	github.com/lib/pq v1.10.8
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/nocodeleaks/quepasa/environment v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/library v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/media v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/rabbitmq v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/whatsapp v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/whatsmeow v0.0.0-00010101000000-000000000000
	github.com/philippseith/signalr v0.6.3
	github.com/sirupsen/logrus v1.9.3
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	go.mau.fi/whatsmeow v0.0.0-20250913213658-6e8bb0a6f77f
	golang.org/x/crypto v0.42.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/beeper/argo-go v1.1.2 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cettoana/go-waveform v0.0.0-20210107122202-35aaec2de427 // indirect
	github.com/elliotchance/orderedmap/v3 v3.1.0 // indirect
	github.com/go-chi/chi/v5 v5.2.2 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/gopxl/beep/v2 v2.1.1 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/gosimple/slug v1.13.1 // indirect
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/hajimehoshi/go-mp3 v0.3.4 // indirect
	github.com/jfreymuth/oggvorbis v1.0.5 // indirect
	github.com/jfreymuth/vorbis v1.0.2 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/klauspost/compress v1.16.6 // indirect
	github.com/mattetti/audio v0.0.0-20240411020228-c5379f9b5b61 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/petermattis/goid v0.0.0-20250813065127-a731cc31b4fe // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	github.com/tcolgate/mp3 v0.0.0-20170426193717-e79c5a46d300 // indirect
	github.com/teivah/onecontext v1.3.0 // indirect
	github.com/vektah/gqlparser/v2 v2.5.30 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.mau.fi/libsignal v0.2.0 // indirect
	go.mau.fi/util v0.9.1-0.20250912114103-419604f95907 // indirect
	golang.org/x/exp v0.0.0-20250813145105-42675adae3e6 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
	nhooyr.io/websocket v1.8.7 // indirect
)

go 1.24.0

toolchain go1.24.2
