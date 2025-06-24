package audio

import (
	"time" // Required for time.Duration
)

// AudioInfo contains basic information about an audio file,
// derived typically from ffprobe results.
type AudioInfo struct {
	Duration   time.Duration
	Channels   int
	SampleRate int
	MIMEType   string
}
