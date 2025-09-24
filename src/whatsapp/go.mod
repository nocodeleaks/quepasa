module github.com/nocodeleaks/quepasa/whatsapp

require github.com/nocodeleaks/quepasa/library v0.0.0-00010101000000-000000000000

require github.com/sirupsen/logrus v1.9.3

require (
	github.com/go-chi/chi/v5 v5.2.2 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/sys v0.36.0 // indirect
)

replace github.com/nocodeleaks/quepasa/whatsapp => ./

replace github.com/nocodeleaks/quepasa/library => ../library

go 1.24.0

toolchain go1.24.2
