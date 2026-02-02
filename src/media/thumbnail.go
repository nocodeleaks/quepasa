package media

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"os/exec"
	"strings"
	"sync"

	// Register image decoders
	_ "image/gif"
	_ "image/png"

	"github.com/nfnt/resize"
	log "github.com/sirupsen/logrus"
)

// ThumbnailConfig holds configuration for thumbnail generation
type ThumbnailConfig struct {
	MaxWidth  uint
	MaxHeight uint
	Quality   int // JPEG quality 1-100
}

// DefaultThumbnailConfig returns default thumbnail settings
// WhatsApp typically uses 72x72 to 320x320 thumbnails
func DefaultThumbnailConfig() ThumbnailConfig {
	return ThumbnailConfig{
		MaxWidth:  320,
		MaxHeight: 320,
		Quality:   75,
	}
}

// Static flags for FFmpeg thumbnail availability
var ffmpegThumbnailAvailable bool
var ffmpegThumbnailOnce sync.Once

// Static flags for Poppler (pdftoppm) availability
var popplerAvailable bool
var popplerOnce sync.Once

// IsFFmpegThumbnailAvailable checks if FFmpeg is available for thumbnail generation
func IsFFmpegThumbnailAvailable() bool {
	ffmpegThumbnailOnce.Do(func() {
		_, err := exec.LookPath("ffmpeg")
		ffmpegThumbnailAvailable = (err == nil)
		if !ffmpegThumbnailAvailable {
			log.Warn("FFmpeg not available for thumbnail generation")
		}
	})
	return ffmpegThumbnailAvailable
}

// IsPopplerAvailable checks if Poppler (pdftoppm) is available for PDF thumbnail generation
func IsPopplerAvailable() bool {
	popplerOnce.Do(func() {
		_, err := exec.LookPath("pdftoppm")
		popplerAvailable = (err == nil)
		if !popplerAvailable {
			log.Warn("Poppler (pdftoppm) not available for PDF thumbnail generation")
		} else {
			log.Debug("Poppler (pdftoppm) available for PDF thumbnail generation")
		}
	})
	return popplerAvailable
}

// GenerateImageThumbnail creates a JPEG thumbnail from image data
func GenerateImageThumbnail(imageData []byte, config ThumbnailConfig) ([]byte, error) {
	if len(imageData) == 0 {
		return nil, fmt.Errorf("empty image data")
	}

	// Decode image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		log.Debugf("Failed to decode image natively (format: %s): %v, trying FFmpeg", format, err)
		// Try FFmpeg as fallback
		return generateThumbnailWithFFmpeg(imageData, "image", config)
	}

	log.Debugf("Decoded image format: %s, size: %dx%d", format, img.Bounds().Dx(), img.Bounds().Dy())

	// Resize image maintaining aspect ratio
	thumbnail := resize.Thumbnail(config.MaxWidth, config.MaxHeight, img, resize.Lanczos3)

	// Encode as JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: config.Quality})
	if err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail as JPEG: %w", err)
	}

	log.Debugf("Generated image thumbnail: %d bytes", buf.Len())
	return buf.Bytes(), nil
}

// GenerateVideoThumbnail creates a JPEG thumbnail from video data
// Extracts frame at 1 second or first frame
func GenerateVideoThumbnail(videoData []byte, config ThumbnailConfig) ([]byte, error) {
	if len(videoData) == 0 {
		return nil, fmt.Errorf("empty video data")
	}

	if !IsFFmpegThumbnailAvailable() {
		return nil, fmt.Errorf("FFmpeg not available for video thumbnail generation")
	}

	return generateThumbnailWithFFmpeg(videoData, "video", config)
}

// GeneratePDFThumbnail creates a JPEG thumbnail from PDF data using Poppler (pdftoppm)
// Renders first page as image
func GeneratePDFThumbnail(pdfData []byte, config ThumbnailConfig) ([]byte, error) {
	if len(pdfData) == 0 {
		return nil, fmt.Errorf("empty PDF data")
	}

	if !IsPopplerAvailable() {
		return nil, fmt.Errorf("Poppler (pdftoppm) not available for PDF thumbnail generation")
	}

	return generatePDFThumbnailWithPoppler(pdfData, config)
}

// generatePDFThumbnailWithPoppler uses Poppler's pdftoppm to generate PDF thumbnails
func generatePDFThumbnailWithPoppler(pdfData []byte, config ThumbnailConfig) ([]byte, error) {
	// Create temporary input file
	inputFile, err := os.CreateTemp("", "thumb-input-*.pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp input file: %w", err)
	}
	defer os.Remove(inputFile.Name())

	if _, err := inputFile.Write(pdfData); err != nil {
		inputFile.Close()
		return nil, fmt.Errorf("failed to write PDF data: %w", err)
	}
	inputFile.Close()

	// Create temporary output file prefix (pdftoppm adds -1.jpg suffix)
	outputPrefix, err := os.CreateTemp("", "thumb-output-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp output prefix: %w", err)
	}
	outputPrefixName := outputPrefix.Name()
	outputPrefix.Close()
	os.Remove(outputPrefixName) // pdftoppm will create the file

	// The actual output file will be outputPrefixName-1.jpg
	outputFileName := outputPrefixName + "-1.jpg"
	defer os.Remove(outputFileName)
	defer os.Remove(outputPrefixName) // cleanup prefix if exists

	// Run pdftoppm command
	// -jpeg: output as JPEG
	// -f 1 -l 1: first page only
	// -scale-to: scale to max dimension
	// -singlefile: don't add page number suffix (but we handle it anyway)
	cmd := exec.Command("pdftoppm",
		"-jpeg",
		"-f", "1",
		"-l", "1",
		"-scale-to", fmt.Sprintf("%d", config.MaxWidth),
		inputFile.Name(),
		outputPrefixName,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Tracef("Executing pdftoppm command: %s", cmd.String())
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("pdftoppm thumbnail generation failed: %w\nstderr: %s", err, stderr.String())
	}

	// Read output file
	thumbData, err := os.ReadFile(outputFileName)
	if err != nil {
		// Try without the -1 suffix (singlefile mode)
		thumbData, err = os.ReadFile(outputPrefixName + ".jpg")
		if err != nil {
			return nil, fmt.Errorf("failed to read PDF thumbnail output: %w", err)
		}
		defer os.Remove(outputPrefixName + ".jpg")
	}

	log.Debugf("Generated PDF thumbnail with Poppler: %d bytes", len(thumbData))
	return thumbData, nil
}

// generateThumbnailWithFFmpeg uses FFmpeg to generate thumbnails for videos and images
func generateThumbnailWithFFmpeg(data []byte, mediaType string, config ThumbnailConfig) ([]byte, error) {
	if !IsFFmpegThumbnailAvailable() {
		return nil, fmt.Errorf("FFmpeg not available")
	}

	// Create temporary input file
	var inputExt string
	switch mediaType {
	case "video":
		inputExt = ".mp4"
	default:
		inputExt = ".bin"
	}

	inputFile, err := os.CreateTemp("", "thumb-input-*"+inputExt)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp input file: %w", err)
	}
	defer os.Remove(inputFile.Name())

	if _, err := inputFile.Write(data); err != nil {
		inputFile.Close()
		return nil, fmt.Errorf("failed to write input data: %w", err)
	}
	inputFile.Close()

	// Create temporary output file
	outputFile, err := os.CreateTemp("", "thumb-output-*.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp output file: %w", err)
	}
	defer os.Remove(outputFile.Name())
	outputFile.Close()

	// Build FFmpeg command based on media type
	var cmd *exec.Cmd
	scaleFilter := fmt.Sprintf("scale='min(%d,iw)':'min(%d,ih)':force_original_aspect_ratio=decrease",
		config.MaxWidth, config.MaxHeight)

	switch mediaType {
	case "video":
		// Extract frame at 1 second (or first frame if video is shorter)
		cmd = exec.Command("ffmpeg",
			"-i", inputFile.Name(),
			"-ss", "00:00:01",
			"-vframes", "1",
			"-vf", scaleFilter,
			"-f", "image2",
			"-vcodec", "mjpeg",
			"-q:v", "5",
			"-y",
			outputFile.Name(),
		)
	default:
		// Generic image conversion
		cmd = exec.Command("ffmpeg",
			"-i", inputFile.Name(),
			"-vf", scaleFilter,
			"-f", "image2",
			"-vcodec", "mjpeg",
			"-q:v", "5",
			"-y",
			outputFile.Name(),
		)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Tracef("Executing FFmpeg thumbnail command: %s", cmd.String())
	err = cmd.Run()
	if err != nil {
		// For video, try without -ss if it failed (very short video)
		if mediaType == "video" {
			log.Debugf("FFmpeg with -ss failed, trying without: %v", err)
			cmd = exec.Command("ffmpeg",
				"-i", inputFile.Name(),
				"-vframes", "1",
				"-vf", scaleFilter,
				"-f", "image2",
				"-vcodec", "mjpeg",
				"-q:v", "5",
				"-y",
				outputFile.Name(),
			)
			stderr.Reset()
			cmd.Stderr = &stderr
			err = cmd.Run()
		}
		if err != nil {
			return nil, fmt.Errorf("FFmpeg thumbnail generation failed: %w\nstderr: %s", err, stderr.String())
		}
	}

	// Read output
	thumbData, err := os.ReadFile(outputFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read thumbnail output: %w", err)
	}

	log.Debugf("Generated %s thumbnail with FFmpeg: %d bytes", mediaType, len(thumbData))
	return thumbData, nil
}

// GenerateThumbnailByMime generates a thumbnail based on MIME type
func GenerateThumbnailByMime(data []byte, mimeType string) ([]byte, error) {
	config := DefaultThumbnailConfig()
	mimeType = strings.ToLower(mimeType)

	if strings.HasPrefix(mimeType, "image/") {
		return GenerateImageThumbnail(data, config)
	}

	if strings.HasPrefix(mimeType, "video/") {
		return GenerateVideoThumbnail(data, config)
	}

	// PDF thumbnail generation using Poppler (pdftoppm)
	if mimeType == "application/pdf" {
		return GeneratePDFThumbnail(data, config)
	}

	return nil, fmt.Errorf("unsupported MIME type for thumbnail generation: %s", mimeType)
}

// CanGenerateThumbnail checks if thumbnail generation is supported for a MIME type
func CanGenerateThumbnail(mimeType string) bool {
	mimeType = strings.ToLower(mimeType)

	// Images - always supported (native Go decoding)
	if strings.HasPrefix(mimeType, "image/") {
		return true
	}

	// Videos - require FFmpeg
	if strings.HasPrefix(mimeType, "video/") {
		return IsFFmpegThumbnailAvailable()
	}

	// PDFs - require Poppler (pdftoppm)
	if mimeType == "application/pdf" {
		return IsPopplerAvailable()
	}

	return false
}
