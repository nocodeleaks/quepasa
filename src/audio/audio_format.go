package audio

import (
	"bytes"   // Required for bytes.HasPrefix
	"strings" // Required for strings.ToLower, strings.HasPrefix
)

// AudioFormat represents the detected audio format.
type AudioFormat string

const (
	FormatUnknown AudioFormat = "unknown"
	FormatMP3     AudioFormat = "mp3"
	FormatOGG     AudioFormat = "ogg"
	FormatWAV     AudioFormat = "wav"
	FormatFLAC    AudioFormat = "flac"
	FormatAAC     AudioFormat = "aac"
)

// detectAudioFormat tries to guess the audio format by reading the first bytes.
func detectAudioFormat(data []byte) AudioFormat {
	if len(data) < 4 {
		return FormatUnknown
	}

	// Basic MP3 detection (ID3 tag or MPEG sync word)
	if len(data) >= 3 && data[0] == 0x49 && data[1] == 0x44 && data[2] == 0x33 { // ID3 tag 'ID3'
		return FormatMP3
	}
	if len(data) >= 2 && data[0] == 0xFF && (data[1]&0xF0) == 0xF0 { // MPEG sync word (first two bytes of frame header)
		return FormatMP3
	}
	if bytes.HasPrefix(data, []byte("OggS")) {
		return FormatOGG
	}
	if bytes.HasPrefix(data, []byte("RIFF")) && len(data) >= 12 && bytes.HasPrefix(data[8:], []byte("WAVE")) {
		return FormatWAV
	}
	if bytes.HasPrefix(data, []byte("fLaC")) {
		return FormatFLAC
	}

	return FormatUnknown
}

// IsAudioMIMEType checks if a MIME type string represents an audio format.
func IsAudioMIMEType(mimeType string) bool {
	lowerMimeType := strings.ToLower(mimeType)
	return strings.HasPrefix(lowerMimeType, "audio/")
}
