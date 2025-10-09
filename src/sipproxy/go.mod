module github.com/nocodeleaks/quepasa/sipproxy

go 1.24.0

toolchain go1.24.2

replace github.com/nocodeleaks/quepasa/environment => ../environment

replace github.com/nocodeleaks/quepasa/whatsapp => ../whatsapp

replace github.com/nocodeleaks/quepasa/library => ../library

require (
	github.com/emiago/sipgo v0.33.0
	github.com/huin/goupnp v1.3.0
	github.com/nocodeleaks/quepasa/environment v0.0.0-00010101000000-000000000000
	github.com/pion/stun v0.6.1
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/go-chi/chi/v5 v5.2.2 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.3.2 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/icholy/digest v1.1.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/nocodeleaks/quepasa/library v0.0.0-00010101000000-000000000000 // indirect
	github.com/nocodeleaks/quepasa/whatsapp v0.0.0-00010101000000-000000000000 // indirect
	github.com/pion/dtls/v2 v2.2.7 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/transport/v2 v2.2.1 // indirect
	golang.org/x/crypto v0.42.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
)
