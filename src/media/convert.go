package media

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/nocodeleaks/quepasa/qplog"
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

// ConvertToWebP converts image or video data to WebP format using FFmpeg,
// conforming to WhatsApp sticker requirements (512x512, transparent padding,
// max 10 seconds at 15fps for animated stickers).
//
// Returns converted data, output MIME type ("image/webp" or "video/webp"), and any error.
func ConvertToWebP(data []byte, inputMime string) ([]byte, string, error) {
	if !IsFFMpegAvailable() {
		return nil, "", fmt.Errorf("ffmpeg is not available for WebP conversion: %w", GetInitError())
	}

	inputFile, err := os.CreateTemp("", "sticker-input-*")
	if err != nil {
		return nil, "", fmt.Errorf("error creating temporary input file: %w", err)
	}
	defer os.Remove(inputFile.Name())

	if _, err := inputFile.Write(data); err != nil {
		inputFile.Close()
		return nil, "", fmt.Errorf("error writing data to temporary input file: %w", err)
	}
	if err := inputFile.Close(); err != nil {
		return nil, "", fmt.Errorf("error closing temporary input file: %w", err)
	}

	outputFile, err := os.CreateTemp("", "sticker-output-*.webp")
	if err != nil {
		return nil, "", fmt.Errorf("error creating temporary output file: %w", err)
	}
	defer os.Remove(outputFile.Name())
	outputFile.Close()

	// Determine if this is a video (animated) or image (static) sticker
	isVideo := strings.HasPrefix(inputMime, "video/") ||
		inputMime == "image/gif" ||
		inputMime == "image/apng"

	var cmd *exec.Cmd
	var outputMime string

	if isVideo {
		// Animated sticker: scale to 512x512 with transparent padding,
		// limit to 10 seconds at 15fps, output as animated WebP.
		cmd = exec.Command("ffmpeg",
			"-i", inputFile.Name(),
			"-t", "10", // max 10 seconds
			"-vf", "fps=15,scale=512:512:force_original_aspect_ratio=decrease,pad=512:512:(ow-iw)/2:(oh-ih)/2:color=0x00000000",
			"-loop", "0",
			"-f", "webp",
			"-y",
			outputFile.Name(),
		)
		outputMime = "video/webp"
	} else {
		// Static sticker: scale to 512x512 with transparent padding,
		// output as static WebP.
		cmd = exec.Command("ffmpeg",
			"-i", inputFile.Name(),
			"-vf", "scale=512:512:force_original_aspect_ratio=decrease,pad=512:512:(ow-iw)/2:(oh-ih)/2:color=0x00000000",
			"-vframes", "1",
			"-f", "webp",
			"-y",
			outputFile.Name(),
		)
		outputMime = "image/webp"
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Infof("Executing ffmpeg WebP conversion (animated=%v): %s", isVideo, cmd.String())
	if err := cmd.Run(); err != nil {
		return nil, "", fmt.Errorf("error converting to WebP with ffmpeg: %w\nstderr: %s", err, stderr.String())
	}

	convertedData, err := os.ReadFile(outputFile.Name())
	if err != nil {
		return nil, "", fmt.Errorf("error reading converted WebP file: %w", err)
	}

	if len(convertedData) == 0 {
		return nil, "", fmt.Errorf("converted WebP file is empty")
	}

	log.Infof("Media successfully converted to WebP (%s): %d bytes input -> %d bytes output", outputMime, len(data), len(convertedData))
	return convertedData, outputMime, nil
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
