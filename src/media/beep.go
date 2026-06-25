package media

import (
	"fmt"
	"strings"

	beep "github.com/gopxl/beep/v2"
)

func GetStreamSeek(mimeType string, content []byte) (beep.Streamer, error) {

	mime := mimeType
	if strings.Contains(mime, ";") {
		mime = strings.Split(mime, ";")[0]
	}

	switch mime {
	case "audio/wave":
		return GetStreamSeekFromWav(content)
	case "audio/ogg":
		return GetStreamSeekFromVorbis(content)
	case "audio/mpeg":
		return GetStreamSeekFromMP3(content)
	default:
		return nil, fmt.Errorf("invalid mime type: %s", mimeType)
	}
}

func GetDuration(mimeType string, content []byte) (uint32, error) {

	mime := mimeType
	if strings.Contains(mime, ";") {
		mime = strings.Split(mime, ";")[0]
	}

	switch mime {
	case "audio/wave":
		return GetDurationFromWav(content)
	case "audio/mpeg":
		return GetDurationFromMP3(content)
	default:
		return 0, fmt.Errorf("invalid mime type: %s", mimeType)
	}
}

/*
func GenerateWaveform(mimeType string, content []byte) ([]byte, error) {
	streamer, err := GetStreamSeek(mimeType, content)
	if err != nil {
		return nil, err
	}

	const numSamples = 64
	samples := make([]float64, 0)
	buf := make([][2]float64, 1024)

	// Converting stereo to mono
	for {
		n, ok := streamer.Stream(buf)
		if !ok {
			break
		}
		for i := 0; i < n; i++ {
			sample := (buf[i][0] + buf[i][1]) / 2
			samples = append(samples, sample)
		}

	}

	// Split samples into blocks to generate 64 values
	blockSize := len(samples) / numSamples
	filteredData := make([]float64, numSamples)

	var maxAmplitude float64 = 0
	for i := 0; i < numSamples; i++ {
		start := i * blockSize
		end := start + blockSize
		if end > len(samples) {
			end = len(samples)
		}

		// Calculate the average amplitude in the block
		var sum float64
		for j := start; j < end; j++ {
			sum += math.Abs(samples[j])
		}
		avg := sum / float64(blockSize)

		filteredData[i] = avg
		if avg > maxAmplitude {
			maxAmplitude = avg
		}
	}

	// Normalize data based on maximum value
	normalizedData := make([]byte, numSamples)
	for i := 0; i < numSamples; i++ {
		if maxAmplitude != 0 {
			normalizedData[i] = byte((filteredData[i] / maxAmplitude) * 100)
		} else {
			normalizedData[i] = 0
		}
	}

	return normalizedData, nil
}
*/
