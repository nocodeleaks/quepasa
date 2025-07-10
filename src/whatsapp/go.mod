module github.com/nocodeleaks/quepasa/whatsapp

require github.com/nocodeleaks/quepasa/library v0.0.0-00010101000000-000000000000

require github.com/sirupsen/logrus v1.9.3

require (
	github.com/stretchr/testify v1.10.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
)

replace github.com/nocodeleaks/quepasa/whatsapp => ./

replace github.com/nocodeleaks/quepasa/library => ../library

go 1.23.0

toolchain go1.24.2
