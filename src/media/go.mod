module github.com/nocodeleaks/quepasa/media

replace github.com/nocodeleaks/quepasa/api => ../api

replace github.com/nocodeleaks/quepasa/environment => ../environment

replace github.com/nocodeleaks/quepasa/form => ../form

replace github.com/nocodeleaks/quepasa/media => ../media

replace github.com/nocodeleaks/quepasa/metrics => ../metrics

replace github.com/nocodeleaks/quepasa/models => ../models

replace github.com/nocodeleaks/quepasa/rabbitmq => ../rabbitmq

replace github.com/nocodeleaks/quepasa/signalr => ../signalr

replace github.com/nocodeleaks/quepasa/sipproxy => ../sipproxy

replace github.com/nocodeleaks/quepasa/swagger => ../swagger

replace github.com/nocodeleaks/quepasa/webserver => ../webserver

replace github.com/nocodeleaks/quepasa/whatsmeow => ../whatsmeow

go 1.24.0 // Ou a sua versão do Go

toolchain go1.24.2

require (
	github.com/cettoana/go-waveform v0.0.0-20210107122202-35aaec2de427
	github.com/gopxl/beep/v2 v2.1.1 // A versão que você quer usar
	github.com/mattetti/audio v0.0.0-20240411020228-c5379f9b5b61
	github.com/nocodeleaks/quepasa/whatsapp v0.0.0-00010101000000-000000000000
	github.com/tcolgate/mp3 v0.0.0-20170426193717-e79c5a46d300
)

require github.com/sirupsen/logrus v1.9.3

require (
	github.com/go-chi/chi/v5 v5.2.3 // indirect
	github.com/hajimehoshi/go-mp3 v0.3.4 // indirect
	github.com/jfreymuth/oggvorbis v1.0.5 // indirect
	github.com/jfreymuth/vorbis v1.0.2 // indirect
	github.com/nocodeleaks/quepasa/library v0.0.0-00010101000000-000000000000 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/sys v0.37.0 // indirect
)

replace github.com/nocodeleaks/quepasa/whatsapp => ../whatsapp

replace github.com/nocodeleaks/quepasa/library => ../library

// Adicione esta seção de "replace"
replace github.com/gopxl/beep v1.4.1 => github.com/gopxl/beep/v2 v2.1.1
