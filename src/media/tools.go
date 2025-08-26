package media

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync" // Required for sync.Once
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/wav"
	logrus "github.com/sirupsen/logrus" // Import Logrus and alias it to logrus
)

// logentry is the Logrus instance used for logging within this package.
// This variable MUST be initialized somewhere in your main application
// or in an init() function within another file of the audio package.
// Example in main.go or another audio package file:
//
// package media
//
// import (
//
//	"os"
//	logrus "github.com/sirupsen/logrus"
//
// )
//
// // Initialize logentry in an init function or by passing it.
// // For simplicity in this example, we'll initialize it here.
// var logentry *logrus.Logger = logrus.New()
//
//	func init() {
//	   logentry.SetFormatter(&logrus.JSONFormatter{})
//	   logentry.SetOutput(os.Stdout)
//	   logentry.SetLevel(logrus.InfoLevel) // Set your desired log level
//	}
var logentry *logrus.Logger = logrus.New() // Default initialization. Prefer external configuration.

// Static flags to store FFmpeg and FFprobe availability.
// These are initialized to false and checked once via sync.Once.
var ffmpegAvailable bool
var ffprobeAvailable bool

// initError stores the error if FFmpeg/FFprobe are not found during their first check.
// This error will contain the reason for the unavailability of either tool.
var initError error

// once ensures the availability checks are performed only once.
var ffmpegOnce sync.Once
var ffprobeOnce sync.Once

// IsFFMpegAvailable checks if the ffmpeg executable is available in the system's PATH.
// The check is performed only once and the result is cached.
func IsFFMpegAvailable() bool {
	ffmpegOnce.Do(func() {
		_, err := exec.LookPath("ffmpeg")
		if err != nil {
			ffmpegAvailable = false
			// Store the detailed error for later retrieval if needed
			// If both are missing, initError will contain the first one found.
			if initError == nil { // Only set if no error has been set yet
				initError = fmt.Errorf("ffmpeg not found in PATH: %w", err)
			}
			logentry.Errorf("FFmpeg is not available. Please ensure it's installed and in your system's PATH. Error: %v", err)
		} else {
			ffmpegAvailable = true
			logentry.Infof("FFmpeg found in PATH.")
		}
	})
	return ffmpegAvailable
}

// IsFFProbeAvailable checks if the ffprobe executable is available in the system's PATH.
// The check is performed only once and the result is cached.
func IsFFProbeAvailable() bool {
	ffprobeOnce.Do(func() {
		_, err := exec.LookPath("ffprobe")
		if err != nil {
			ffprobeAvailable = false
			// Store the detailed error for later retrieval if needed.
			// Prioritize existing initError if ffmpeg was already missing.
			if initError == nil { // Only set if no error has been set yet
				initError = fmt.Errorf("ffprobe not found in PATH: %w", err)
			}
			logentry.Errorf("FFprobe is not available. Please ensure it's installed and in your system's PATH. Error: %v", err)
		} else {
			ffprobeAvailable = true
			logentry.Infof("FFprobe found in PATH.")
		}
	})
	return ffprobeAvailable
}

// AreAudioToolsAvailable checks if both ffmpeg and ffprobe executables are available.
// It calls IsFFMpegAvailable and IsFFProbeAvailable internally, caching their results.
// Returns true if both are available, false otherwise.
func AreAudioToolsAvailable() bool {
	return IsFFMpegAvailable() && IsFFProbeAvailable()
}

// GetInitError returns the error encountered during the initial availability check
// of FFmpeg or FFprobe, if any. Returns nil if both are available.
// It ensures the checks have been performed by calling AreAudioToolsAvailable.
func GetInitError() error {
	AreAudioToolsAvailable() // Ensure both checks have run
	return initError
}

// GetAudioInfoFromBytes retrieves audio information
// from a byte slice, using ffprobe and a temporary file.
func GetAudioInfoFromBytes(audioData []byte) (*AudioInfo, error) {
	// Check if ffprobe is available before proceeding
	if !IsFFProbeAvailable() {
		// Return the existing error from the initial check
		return nil, fmt.Errorf("ffprobe is not available: %w", GetInitError())
	}

	tmpfile, err := os.CreateTemp("", "audio-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary file: %w", err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	if _, err := tmpfile.Write(audioData); err != nil {
		return nil, fmt.Errorf("error writing data to temporary file: %w", err)
	}
	if err := tmpfile.Sync(); err != nil {
		return nil, fmt.Errorf("error syncing temporary file: %w", err)
	}
	if err := tmpfile.Close(); err != nil {
		return nil, fmt.Errorf("error closing temporary file: %w", err)
	}

	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration,format_name,filename,bit_rate:stream=channels,sample_rate,codec_type,codec_name",
		"-of", "json",
		tmpfile.Name(),
	)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	logentry.Infof("Executing ffprobe command: %s", cmd.String())
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error executing ffprobe: %w\nstderr: %s", err, stderr.String())
	}

	var result FFProbeResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling ffprobe JSON output: %w\nOutput: %s", err, out.String())
	}

	audioInfo := &AudioInfo{}

	durationStr := result.Format.Duration
	if durationStr == "" {
		return nil, fmt.Errorf("duration not found in ffprobe output for '%s'. Output: %s", tmpfile.Name(), out.String())
	}
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return nil, fmt.Errorf("error converting duration '%s' to float: %w", durationStr, err)
	}
	audioInfo.Duration = time.Duration(durationFloat * float64(time.Second))

	foundAudioStream := false
	for _, stream := range result.Streams {
		if stream.CodecType == "audio" {
			audioInfo.Channels = stream.Channels
			sampleRateInt, parseErr := strconv.Atoi(stream.SampleRate)
			if parseErr != nil {
				logentry.Warnf("Warning: Could not convert SampleRate '%s' to int: %v", stream.SampleRate, parseErr)
				audioInfo.SampleRate = 0
			} else {
				audioInfo.SampleRate = sampleRateInt
			}
			foundAudioStream = true
			break
		}
	}
	if !foundAudioStream {
		return nil, fmt.Errorf("no audio stream found in the file")
	}

	switch result.Format.FormatName {
	case "ogg":
		if len(result.Streams) > 0 && result.Streams[0].CodecName == "opus" {
			audioInfo.MIMEType = "audio/opus"
		} else {
			audioInfo.MIMEType = "audio/ogg"
		}
	case "mp3":
		audioInfo.MIMEType = "audio/mpeg"
	case "wav":
		audioInfo.MIMEType = "audio/wav"
	case "flac":
		audioInfo.MIMEType = "audio/flac"
	case "aac":
		audioInfo.MIMEType = "audio/aac"
	default:
		audioInfo.MIMEType = fmt.Sprintf("audio/%s", result.Format.FormatName)
	}

	return audioInfo, nil
}

// transcodeToWAV uses ffmpeg to convert audio data to WAV format.
func transcodeToWAV(audioData []byte, inputFormat AudioFormat) ([]byte, error) {
	// Check if ffmpeg is available before proceeding
	if !IsFFMpegAvailable() {
		return nil, fmt.Errorf("ffmpeg is not available: %w", GetInitError())
	}

	inputFile, err := os.CreateTemp("", fmt.Sprintf("input-*.%s", inputFormat))
	if err != nil {
		return nil, fmt.Errorf("error creating temporary input file for transcoding: %w", err)
	}
	defer os.Remove(inputFile.Name())
	defer inputFile.Close()

	if _, err := inputFile.Write(audioData); err != nil {
		return nil, fmt.Errorf("error writing data to temporary input file: %w", err)
	}
	if err := inputFile.Sync(); err != nil {
		return nil, fmt.Errorf("error syncing temporary input file: %w", err)
	}
	if err := inputFile.Close(); err != nil {
		return nil, fmt.Errorf("error closing temporary input file: %w", err)
	}

	outputFile, err := os.CreateTemp("", "output-*.wav")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary output file for transcoding: %w", err)
	}
	defer os.Remove(outputFile.Name())
	defer outputFile.Close()

	cmd := exec.Command("ffmpeg",
		"-i", inputFile.Name(),
		"-f", "wav",
		"-y", // Overwrite output file without asking
		outputFile.Name(),
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	logentry.Infof("Executing ffmpeg transcoding: %s", cmd.String())
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error transcoding audio with ffmpeg: %w\nstderr: %s", err, stderr.String())
	}

	transcodedData, err := os.ReadFile(outputFile.Name())
	if err != nil {
		return nil, fmt.Errorf("error reading transcoded WAV file: %w", err)
	}

	return transcodedData, nil
}

// GenerateWaveform generates a 64-byte waveform from an audio byte slice.
// It transcodes OGG, FLAC, and unsupported MP3s (MPEG 2.5) to WAV.
func GenerateWaveform(audioData []byte) ([]byte, error) {
	// Check if both ffmpeg and ffprobe are available.
	if !AreAudioToolsAvailable() {
		return nil, fmt.Errorf("required audio tools (ffmpeg/ffprobe) are not available. Details: %w", GetInitError())
	}

	const numSamples = 64

	originalAudioFormat := detectAudioFormat(audioData)
	currentAudioFormat := originalAudioFormat
	currentAudioData := audioData

	var streamer beep.StreamSeekCloser
	var format beep.Format
	var decodeErr error

	needsTranscoding := false

	switch currentAudioFormat {
	case FormatMP3:
		audioReader := io.NopCloser(bytes.NewReader(currentAudioData))
		streamer, format, decodeErr = mp3.Decode(audioReader)
		if decodeErr != nil {
			if strings.Contains(decodeErr.Error(), "MPEG version 2.5 is not supported") {
				logentry.Infof("MP3 with MPEG version 2.5 is not supported by Beep. Attempting to transcode to WAV.")
				needsTranscoding = true
			} else {
				if streamer != nil {
					streamer.Close()
				}
				return nil, fmt.Errorf("failed to decode MP3: %w", decodeErr)
			}
		}
	case FormatWAV:
		audioReader := io.NopCloser(bytes.NewReader(currentAudioData))
		streamer, format, decodeErr = wav.Decode(audioReader)
	case FormatOGG, FormatFLAC:
		logentry.Infof("Format %s not directly decodable by Beep. Transcoding to WAV.", currentAudioFormat)
		needsTranscoding = true
	default:
		return nil, fmt.Errorf("audio format not supported for waveform generation: %s", currentAudioFormat)
	}

	if needsTranscoding {
		var err error
		currentAudioData, err = transcodeToWAV(audioData, originalAudioFormat)
		if err != nil {
			return nil, fmt.Errorf("failed to transcode %s to WAV: %w", originalAudioFormat, err)
		}
		currentAudioFormat = FormatWAV
		audioReader := io.NopCloser(bytes.NewReader(currentAudioData))
		streamer, format, decodeErr = wav.Decode(audioReader)
	}

	if decodeErr != nil {
		if streamer != nil {
			streamer.Close()
		}
		return nil, fmt.Errorf("final failure to decode audio after transcoding (%s): %w", currentAudioFormat, decodeErr)
	}
	defer streamer.Close()

	resampledStreamer := beep.Resample(4, format.SampleRate, beep.SampleRate(44100), streamer)

	samplesForWaveform := make([]float64, 0)
	bufferForWaveform := make([][2]float64, beep.SampleRate(44100).N(time.Second/10))

	for {
		n, ok := resampledStreamer.Stream(bufferForWaveform)
		if !ok {
			break
		}

		for i := 0; i < n; i++ {
			sample := (bufferForWaveform[i][0] + bufferForWaveform[i][1]) / 2
			samplesForWaveform = append(samplesForWaveform, math.Abs(sample))
		}
	}

	if len(samplesForWaveform) == 0 {
		return make([]byte, numSamples), nil
	}

	blockSize := len(samplesForWaveform) / numSamples
	if blockSize == 0 {
		blockSize = 1
	}

	filteredData := make([]float64, numSamples)
	var maxAmplitudeForWaveformScaling float64 = 0.001

	for i := 0; i < numSamples; i++ {
		start := i * blockSize
		end := start + blockSize
		if end > len(samplesForWaveform) {
			end = len(samplesForWaveform)
		}

		blockMax := 0.0
		for j := start; j < end; j++ {
			if samplesForWaveform[j] > blockMax {
				blockMax = samplesForWaveform[j]
			}
		}
		filteredData[i] = blockMax
		if blockMax > maxAmplitudeForWaveformScaling {
			maxAmplitudeForWaveformScaling = blockMax
		}
	}

	waveform := make([]byte, numSamples)
	for i := 0; i < numSamples; i++ {
		scaledValue := (filteredData[i] / maxAmplitudeForWaveformScaling) * 100.0

		if scaledValue > 100.0 {
			scaledValue = 100.0
		}
		if scaledValue < 0.0 {
			scaledValue = 0.0
		}
		waveform[i] = byte(math.Round(scaledValue))
	}

	return waveform, nil
}
