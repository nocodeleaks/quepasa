package media

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// ConvertToOggOpus converts audio data (any format supported by ffmpeg) to OGG Opus format.
// This is used to convert MP3 and other audio formats to the format required for
// WhatsApp PTT (Push-to-Talk) voice notes.
// Returns the converted OGG Opus audio bytes, or an error if conversion fails.
func ConvertToOggOpus(audioData []byte) ([]byte, error) {
	if !IsFFMpegAvailable() {
		return nil, fmt.Errorf("ffmpeg is not available for audio conversion: %w", GetInitError())
	}

	// Create temporary input file
	inputFile, err := os.CreateTemp("", "audio-input-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary input file: %w", err)
	}
	defer os.Remove(inputFile.Name())

	if _, err := inputFile.Write(audioData); err != nil {
		inputFile.Close()
		return nil, fmt.Errorf("error writing data to temporary input file: %w", err)
	}
	if err := inputFile.Close(); err != nil {
		return nil, fmt.Errorf("error closing temporary input file: %w", err)
	}

	// Create temporary output file
	outputFile, err := os.CreateTemp("", "audio-output-*.ogg")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary output file: %w", err)
	}
	defer os.Remove(outputFile.Name())
	outputFile.Close()

	// Convert using ffmpeg: input → OGG with Opus codec, mono, 48kHz (WhatsApp standard)
	cmd := exec.Command("ffmpeg",
		"-i", inputFile.Name(),
		"-c:a", "libopus", // Opus codec
		"-b:a", "64k", // Bitrate suitable for voice
		"-ac", "1", // Mono channel (voice notes are mono)
		"-ar", "48000", // 48kHz sample rate (Opus standard)
		"-vn",       // No video
		"-f", "ogg", // OGG container
		"-y", // Overwrite output
		outputFile.Name(),
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Infof("Executing ffmpeg audio conversion to OGG Opus: %s", cmd.String())
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error converting audio to OGG Opus with ffmpeg: %w\nstderr: %s", err, stderr.String())
	}

	// Read converted file
	convertedData, err := os.ReadFile(outputFile.Name())
	if err != nil {
		return nil, fmt.Errorf("error reading converted OGG Opus file: %w", err)
	}

	if len(convertedData) == 0 {
		return nil, fmt.Errorf("converted OGG Opus file is empty")
	}

	log.Infof("Audio successfully converted to OGG Opus: %d bytes input → %d bytes output", len(audioData), len(convertedData))
	return convertedData, nil
}

// ShouldConvertToPTT checks if the given MIME type is an audio format that should be
// automatically converted to OGG Opus for PTT (voice note) delivery.
// This covers non-PTT audio formats like MP3, MP4 audio, AAC, OGA, etc.
// WAV/Wave formats are excluded because they are already PTT-compatible
// (handled by the existing PTT-compatible path without ffmpeg).
func ShouldConvertToPTT(mimeType string) bool {
	switch mimeType {
	case "audio/mpeg", "audio/mp3", "audio/x-mpeg-3", "audio/mpeg3",
		"audio/mp4", "audio/aac",
		"audio/oga", "audio/ogx":
		return true
	default:
		return false
	}
}
