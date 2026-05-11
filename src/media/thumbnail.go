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

	_ "image/gif"
	_ "image/png"

	"github.com/nfnt/resize"
	log "github.com/sirupsen/logrus"
)

// ThumbnailConfig holds configuration for thumbnail generation.
type ThumbnailConfig struct {
	MaxWidth  uint
	MaxHeight uint
	Quality   int // JPEG quality 1-100
	MaxBytes  int // Maximum thumbnail size in bytes (0 = no limit)
}

// DefaultThumbnailConfig returns default thumbnail settings.
// WhatsApp has a practical limit around 10KB for JPEGThumbnail.
func DefaultThumbnailConfig() ThumbnailConfig {
	return ThumbnailConfig{
		MaxWidth:  200,
		MaxHeight: 200,
		Quality:   70,
		MaxBytes:  10240,
	}
}

var ffmpegThumbnailAvailable bool
var ffmpegThumbnailOnce sync.Once

var popplerAvailable bool
var popplerOnce sync.Once

// IsFFmpegThumbnailAvailable checks if FFmpeg is available for thumbnail generation.
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

// IsPopplerAvailable checks if Poppler (pdftoppm) is available for PDF thumbnail generation.
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

// GenerateImageThumbnail creates a JPEG thumbnail from image data.
func GenerateImageThumbnail(imageData []byte, config ThumbnailConfig) ([]byte, error) {
	if len(imageData) == 0 {
		return nil, fmt.Errorf("empty image data")
	}

	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		log.Debugf("Failed to decode image natively (format: %s): %v, trying FFmpeg", format, err)
		return generateThumbnailWithFFmpeg(imageData, "image", config)
	}

	thumbnail := resize.Thumbnail(config.MaxWidth, config.MaxHeight, img, resize.Lanczos3)

	quality := config.Quality
	maxBytes := config.MaxBytes
	if maxBytes <= 0 {
		maxBytes = 10240
	}

	var buf bytes.Buffer
	for quality >= 30 {
		buf.Reset()
		err = jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: quality})
		if err != nil {
			return nil, fmt.Errorf("failed to encode thumbnail as JPEG: %w", err)
		}

		if buf.Len() <= maxBytes {
			break
		}

		quality -= 10
	}

	if buf.Len() > maxBytes {
		smallerConfig := config
		smallerConfig.MaxWidth = config.MaxWidth * 3 / 4
		smallerConfig.MaxHeight = config.MaxHeight * 3 / 4
		thumbnail = resize.Thumbnail(smallerConfig.MaxWidth, smallerConfig.MaxHeight, img, resize.Lanczos3)

		buf.Reset()
		err = jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 60})
		if err != nil {
			return nil, fmt.Errorf("failed to encode smaller thumbnail as JPEG: %w", err)
		}
	}

	return buf.Bytes(), nil
}

// GenerateVideoThumbnail creates a JPEG thumbnail from video data.
func GenerateVideoThumbnail(videoData []byte, config ThumbnailConfig) ([]byte, error) {
	if len(videoData) == 0 {
		return nil, fmt.Errorf("empty video data")
	}

	if !IsFFmpegThumbnailAvailable() {
		return nil, fmt.Errorf("FFmpeg not available for video thumbnail generation")
	}

	return generateThumbnailWithFFmpeg(videoData, "video", config)
}

// GeneratePDFThumbnail creates a JPEG thumbnail from PDF data using Poppler (pdftoppm).
func GeneratePDFThumbnail(pdfData []byte, config ThumbnailConfig) ([]byte, error) {
	if len(pdfData) == 0 {
		return nil, fmt.Errorf("empty PDF data")
	}

	if !IsPopplerAvailable() {
		return nil, fmt.Errorf("Poppler (pdftoppm) not available for PDF thumbnail generation")
	}

	return generatePDFThumbnailWithPoppler(pdfData, config)
}

func generatePDFThumbnailWithPoppler(pdfData []byte, config ThumbnailConfig) ([]byte, error) {
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

	outputPrefix, err := os.CreateTemp("", "thumb-output-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp output prefix: %w", err)
	}
	outputPrefixName := outputPrefix.Name()
	outputPrefix.Close()
	os.Remove(outputPrefixName)

	outputFileName := outputPrefixName + "-1.jpg"
	defer os.Remove(outputFileName)
	defer os.Remove(outputPrefixName)

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

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("pdftoppm thumbnail generation failed: %w\nstderr: %s", err, stderr.String())
	}

	thumbData, err := os.ReadFile(outputFileName)
	if err != nil {
		thumbData, err = os.ReadFile(outputPrefixName + ".jpg")
		if err != nil {
			return nil, fmt.Errorf("failed to read PDF thumbnail output: %w", err)
		}
		defer os.Remove(outputPrefixName + ".jpg")
	}

	return thumbData, nil
}

func generateThumbnailWithFFmpeg(data []byte, mediaType string, config ThumbnailConfig) ([]byte, error) {
	if !IsFFmpegThumbnailAvailable() {
		return nil, fmt.Errorf("FFmpeg not available")
	}

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

	outputFile, err := os.CreateTemp("", "thumb-output-*.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp output file: %w", err)
	}
	defer os.Remove(outputFile.Name())
	outputFile.Close()

	scaleFilter := fmt.Sprintf("scale='min(%d,iw)':'min(%d,ih)':force_original_aspect_ratio=decrease",
		config.MaxWidth, config.MaxHeight)

	var cmd *exec.Cmd
	switch mediaType {
	case "video":
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

	err = cmd.Run()
	if err != nil {
		if mediaType == "video" {
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

	thumbData, err := os.ReadFile(outputFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read thumbnail output: %w", err)
	}

	return thumbData, nil
}

// GenerateThumbnailByMime generates a thumbnail based on MIME type.
func GenerateThumbnailByMime(data []byte, mimeType string) ([]byte, error) {
	config := DefaultThumbnailConfig()
	mimeType = strings.ToLower(mimeType)

	if strings.HasPrefix(mimeType, "image/") {
		return GenerateImageThumbnail(data, config)
	}

	if strings.HasPrefix(mimeType, "video/") {
		return GenerateVideoThumbnail(data, config)
	}

	if mimeType == "application/pdf" {
		return GeneratePDFThumbnail(data, config)
	}

	return nil, fmt.Errorf("unsupported MIME type for thumbnail generation: %s", mimeType)
}

// CanGenerateThumbnail checks if thumbnail generation is supported for a MIME type.
func CanGenerateThumbnail(mimeType string) bool {
	mimeType = strings.ToLower(mimeType)

	if strings.HasPrefix(mimeType, "image/") {
		return true
	}

	if strings.HasPrefix(mimeType, "video/") {
		return IsFFmpegThumbnailAvailable()
	}

	if mimeType == "application/pdf" {
		return IsPopplerAvailable()
	}

	return false
}
