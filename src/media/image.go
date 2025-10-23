package media

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Static flags to store FFmpeg availability.
var ffmpegImageAvailable bool
var ffmpegImageOnce sync.Once
var ffmpegImageError error

// IsFFmpegImageAvailable checks if the ffmpeg executable is available for image conversion.
// The check is performed only once and the result is cached.
func IsFFmpegImageAvailable() bool {
	ffmpegImageOnce.Do(func() {
		_, err := exec.LookPath("ffmpeg")
		if err != nil {
			ffmpegImageAvailable = false
			ffmpegImageError = fmt.Errorf("ffmpeg not found in PATH: %w", err)
			log.Errorf("FFmpeg is not available for image conversion. Please ensure it's installed and in your system's PATH. Error: %v", err)
		} else {
			ffmpegImageAvailable = true
			log.Tracef("FFmpeg found in PATH for image conversion.")
		}
	})
	return ffmpegImageAvailable
}

// GetFFmpegImageError returns the error encountered during the initial availability check of FFmpeg.
func GetFFmpegImageError() error {
	IsFFmpegImageAvailable() // Ensure the check has run
	return ffmpegImageError
}

// ConvertPngToJpg converts a PNG image to JPG format using FFmpeg.
// Returns the converted JPG image as bytes and the new MIME type.
func ConvertPngToJpg(pngData []byte) (jpgData []byte, newMime string, err error) {
	// Check if ffmpeg is available before proceeding
	if !IsFFmpegImageAvailable() {
		return nil, "", fmt.Errorf("ffmpeg is not available for image conversion: %w", GetFFmpegImageError())
	}

	// Create temporary input file
	inputFile, err := os.CreateTemp("", "input-*.png")
	if err != nil {
		return nil, "", fmt.Errorf("error creating temporary input file for PNG conversion: %w", err)
	}
	defer os.Remove(inputFile.Name())
	defer inputFile.Close()

	// Write PNG data to input file
	if _, err := inputFile.Write(pngData); err != nil {
		return nil, "", fmt.Errorf("error writing PNG data to temporary input file: %w", err)
	}
	if err := inputFile.Sync(); err != nil {
		return nil, "", fmt.Errorf("error syncing temporary input file: %w", err)
	}
	if err := inputFile.Close(); err != nil {
		return nil, "", fmt.Errorf("error closing temporary input file: %w", err)
	}

	// Create temporary output file
	outputFile, err := os.CreateTemp("", "output-*.jpg")
	if err != nil {
		return nil, "", fmt.Errorf("error creating temporary output file for JPG conversion: %w", err)
	}
	defer os.Remove(outputFile.Name())
	defer outputFile.Close()

	// Run FFmpeg command to convert PNG to JPG
	// Using high quality settings
	cmd := exec.Command("ffmpeg",
		"-i", inputFile.Name(),
		"-f", "image2",
		"-vcodec", "mjpeg",
		"-pix_fmt", "yuvj420p",
		"-q:v", "2", // High quality (1-31, lower is better)
		"-y", // Overwrite output file without asking
		outputFile.Name(),
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Tracef("Executing FFmpeg PNG to JPG conversion: %s", cmd.String())
	err = cmd.Run()
	if err != nil {
		return nil, "", fmt.Errorf("error converting PNG to JPG with ffmpeg: %w\nstderr: %s", err, stderr.String())
	}

	// Read converted JPG data
	jpgData, err = os.ReadFile(outputFile.Name())
	if err != nil {
		return nil, "", fmt.Errorf("error reading converted JPG file: %w", err)
	}

	log.Debugf("Successfully converted PNG to JPG. Original size: %d bytes, converted size: %d bytes", len(pngData), len(jpgData))

	return jpgData, "image/jpeg", nil
}

// ShouldConvertImage checks if an image should be converted based on its MIME type and filename.
func ShouldConvertImage(mimeType, filename string) bool {
	// First check MIME type - this is the most reliable indicator
	lowerMimeType := strings.ToLower(mimeType)

	// If MIME type is definitely an image/png, convert it
	if lowerMimeType == "image/png" {
		return true
	}

	// If MIME type indicates it's another format (audio, video, etc.), don't convert
	// even if filename has .png extension
	if strings.HasPrefix(lowerMimeType, "audio/") ||
		strings.HasPrefix(lowerMimeType, "video/") ||
		strings.HasPrefix(lowerMimeType, "application/") ||
		strings.HasPrefix(lowerMimeType, "text/") {
		return false
	}

	// If MIME type is empty or generic, check file extension as fallback
	if filename != "" {
		ext := strings.ToLower(filepath.Ext(filename))
		return ext == ".png"
	}

	return false
}
