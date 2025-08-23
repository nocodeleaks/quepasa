module github.com/nocodeleaks/quepasa/environment

go 1.23.2

require (
	github.com/nocodeleaks/quepasa/library v0.0.0-00010101000000-000000000000
	github.com/nocodeleaks/quepasa/whatsapp v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/go-chi/chi/v5 v5.2.2 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	golang.org/x/sys v0.33.0 // indirect
)

replace github.com/nocodeleaks/quepasa/library => ../library

replace github.com/nocodeleaks/quepasa/whatsapp => ../whatsapp
