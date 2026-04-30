module github.com/nocodeleaks/quepasa/runtime

go 1.25.0

require (
	github.com/nocodeleaks/quepasa/dispatch v0.0.0
	github.com/nocodeleaks/quepasa/models v0.0.0
	github.com/nocodeleaks/quepasa/whatsapp v0.0.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/beeper/argo-go v1.1.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cettoana/go-waveform v0.0.0-20210107122202-35aaec2de427 // indirect
	github.com/coder/websocket v1.8.14 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/elliotchance/orderedmap/v3 v3.1.0 // indirect
	github.com/go-chi/chi/v5 v5.2.3 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gopxl/beep/v2 v2.1.1 // indirect
	github.com/gosimple/slug v1.13.1 // indirect
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/hajimehoshi/go-mp3 v0.3.4 // indirect
	github.com/jfreymuth/oggvorbis v1.0.5 // indirect
	github.com/jfreymuth/vorbis v1.0.2 // indirect
	github.com/jmoiron/sqlx v1.3.5 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/joncalhoun/migrate v0.0.2 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mattetti/audio v0.0.0-20240411020228-c5379f9b5b61 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646 // indirect
	github.com/nocodeleaks/quepasa/cache v0.0.0-00010101000000-000000000000 // indirect
	github.com/nocodeleaks/quepasa/environment v0.0.0-00010101000000-000000000000 // indirect
	github.com/nocodeleaks/quepasa/events v0.0.0 // indirect
	github.com/nocodeleaks/quepasa/library v0.0.0 // indirect
	github.com/nocodeleaks/quepasa/media v0.0.0-00010101000000-000000000000 // indirect
	github.com/nocodeleaks/quepasa/metrics v0.0.0-00010101000000-000000000000 // indirect
	github.com/nocodeleaks/quepasa/rabbitmq v0.0.0-00010101000000-000000000000 // indirect
	github.com/nocodeleaks/quepasa/webserver v0.0.0-00010101000000-000000000000 // indirect
	github.com/nocodeleaks/quepasa/whatsmeow v0.0.0-00010101000000-000000000000 // indirect
	github.com/petermattis/goid v0.0.0-20260113132338-7c7de50cc741 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.16.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
	github.com/redis/go-redis/v9 v9.7.1 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e // indirect
	github.com/tcolgate/mp3 v0.0.0-20170426193717-e79c5a46d300 // indirect
	github.com/vektah/gqlparser/v2 v2.5.30 // indirect
	go.mau.fi/libsignal v0.2.1 // indirect
	go.mau.fi/util v0.9.6 // indirect
	go.mau.fi/whatsmeow v0.0.0-20260219150138-7ae702b1eed4 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/exp v0.0.0-20260212183809-81e46e3db34a // indirect
	golang.org/x/net v0.50.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace (
	github.com/nocodeleaks/quepasa/cache => ../cache
	github.com/nocodeleaks/quepasa/dispatch => ../dispatch
	github.com/nocodeleaks/quepasa/environment => ../environment
	github.com/nocodeleaks/quepasa/events => ../events
	github.com/nocodeleaks/quepasa/library => ../library
	github.com/nocodeleaks/quepasa/media => ../media
	github.com/nocodeleaks/quepasa/metrics => ../metrics
	github.com/nocodeleaks/quepasa/models => ../models
	github.com/nocodeleaks/quepasa/rabbitmq => ../rabbitmq
	github.com/nocodeleaks/quepasa/webserver => ../webserver
	github.com/nocodeleaks/quepasa/whatsapp => ../whatsapp
	github.com/nocodeleaks/quepasa/whatsmeow => ../whatsmeow
)
