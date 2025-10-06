module github.com/nocodeleaks/quepasa/environment

go 1.24.0

toolchain go1.24.2

require (
	github.com/joho/godotenv v1.5.1
	github.com/nocodeleaks/quepasa/library v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/whatsapp v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/go-chi/chi/v5 v5.2.3 // indirect
	golang.org/x/sys v0.36.0 // indirect
)

replace github.com/nocodeleaks/quepasa/library => ../library

replace github.com/nocodeleaks/quepasa/whatsapp => ../whatsapp
