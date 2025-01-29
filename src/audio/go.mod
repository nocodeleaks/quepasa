module github.com/nocodeleaks/quepasa/audio

replace github.com/nocodeleaks/quepasa/library => ../library

replace github.com/nocodeleaks/quepasa/whatsapp => ../whatsapp

require (
	github.com/cettoana/go-waveform v0.0.0-20210107122202-35aaec2de427
	github.com/gopxl/beep/v2 v2.1.0
	github.com/mattetti/audio v0.0.0-20240411020228-c5379f9b5b61
	github.com/nocodeleaks/quepasa/whatsapp v0.0.0-00010101000000-000000000000
	github.com/tcolgate/mp3 v0.0.0-20170426193717-e79c5a46d300
)

require (
	github.com/hajimehoshi/go-mp3 v0.3.4 // indirect
	github.com/jfreymuth/oggvorbis v1.0.5 // indirect
	github.com/jfreymuth/vorbis v1.0.2 // indirect
	github.com/nocodeleaks/quepasa/library v0.0.0-00010101000000-000000000000 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/sys v0.18.0 // indirect
)

go 1.22.0
