package audio

import (
	"fmt"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type AudioDetails struct {
	Content       []byte
	Mimetype      string
	Channels      uint16
	SampleRate    uint32
	BitsPerSample uint16
	Format        string
	DataChuckSize uint32
	Duration      uint32
	WaveForm      []byte

	Debug []string `json:"debug,omitempty"`
}

func GetAudioDetails(attach *whatsapp.WhatsappAttachment) (debug []string) {

	if attach == nil {
		debug = append(debug, "[error][GetAudioDetails] nil attachment")
		return
	}

	waveformDebug := EnsureWaveForm(attach)
	debug = append(debug, waveformDebug...)

	durationDebug := EnsureDuration(attach)
	debug = append(debug, durationDebug...)

	return
}

func EnsureWaveForm(attach *whatsapp.WhatsappAttachment) (debug []string) {
	if len(attach.WaveForm) > 0 {
		debug = append(debug, "[trace][EnsureWaveForm] already has waveform")
		return
	}

	content := attach.GetContent()
	if content == nil {
		debug = append(debug, "[error][EnsureWaveForm] nil attachment content")
		return
	}

	waveform, err := GenerateWaveform(attach.Mimetype, *content)
	if err != nil {
		debug = append(debug, fmt.Sprintf("[error][EnsureWaveForm] %s", err.Error()))
		return
	}

	attach.WaveForm = waveform
	debug = append(debug, "[debug][EnsureWaveForm] success to generate waveform")
	return
}

func EnsureDuration(attach *whatsapp.WhatsappAttachment) (debug []string) {
	if attach.Seconds > 0 {
		debug = append(debug, "[trace][EnsureDuration] already has duration")
		return
	}

	content := attach.GetContent()
	if content == nil {
		debug = append(debug, "[error][EnsureDuration] nil attachment content")
		return
	}

	duration, err := GetDuration(attach.Mimetype, *content)
	if err != nil {
		debug = append(debug, fmt.Sprintf("[error][EnsureDuration] %s", err.Error()))
		return
	}

	attach.Seconds = duration
	debug = append(debug, fmt.Sprintf("[debug][EnsureDuration] success to get audio duration: %v seconds", duration))
	return
}
