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

	// Convert using ffmpeg: input -> OGG with Opus codec, mono, 48kHz (WhatsApp standard)
	cmd := exec.Command("ffmpeg",
		"-i", inputFile.Name(),
		"-c:a", "libopus",
		"-b:a", "64k",
		"-ac", "1",
		"-ar", "48000",
		"-vn",
		"-f", "ogg",
		"-y",
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

	log.Infof("Audio successfully converted to OGG Opus: %d bytes input -> %d bytes output", len(audioData), len(convertedData))
	return convertedData, nil
}

// ConvertToWebP converts any image or video to WebP using FFmpeg.
// isVideo=true produces animated WebP (video/webp); false produces static WebP (image/webp).
// WhatsApp sticker requirements: 512x512px, max 10s/15fps for animated, transparent padding.
func ConvertToWebP(data []byte, isVideo bool) ([]byte, string, error) {
	if !IsFFMpegAvailable() {
		return nil, "", fmt.Errorf("ffmpeg not available: %w", GetInitError())
	}

	in, err := os.CreateTemp("", "sticker-in-*")
	if err != nil {
		return nil, "", err
	}
	defer os.Remove(in.Name())
	if _, err = in.Write(data); err != nil {
		in.Close()
		return nil, "", err
	}
	in.Close()

	out, err := os.CreateTemp("", "sticker-out-*.webp")
	if err != nil {
		return nil, "", err
	}
	defer os.Remove(out.Name())
	out.Close()

	var args []string
	if isVideo {
		args = []string{
			"-i", in.Name(),
			"-t", "10",
			"-vf", "fps=15,scale=512:512:force_original_aspect_ratio=decrease,pad=512:512:(ow-iw)/2:(oh-ih)/2:color=0x00000000",
			"-loop", "0", "-compression_level", "6", "-y", out.Name(),
		}
	} else {
		args = []string{
			"-i", in.Name(),
			"-vf", "scale=512:512:force_original_aspect_ratio=decrease,pad=512:512:(ow-iw)/2:(oh-ih)/2:color=0x00000000",
			"-vframes", "1", "-y", out.Name(),
		}
	}

	var stderr bytes.Buffer
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		return nil, "", fmt.Errorf("ffmpeg WebP: %w\n%s", err, stderr.String())
	}

	result, err := os.ReadFile(out.Name())
	if err != nil || len(result) == 0 {
		return nil, "", fmt.Errorf("empty WebP output")
	}

	mime := "image/webp"
	if isVideo {
		mime = "video/webp"
	}
	return result, mime, nil
}

// ShouldConvertToPTT checks if the given MIME type is an audio format that should be
// automatically converted to OGG Opus for PTT (voice note) delivery.
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
