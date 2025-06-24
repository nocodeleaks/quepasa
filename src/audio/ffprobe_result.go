package audio

// FFProbeResult struct to unmarshal ffprobe JSON output.
// It contains format-level and stream-level information from ffprobe.
type FFProbeResult struct {
	Format struct {
		Filename   string `json:"filename"`
		Duration   string `json:"duration"`
		FormatName string `json:"format_name"`
		BitRate    string `json:"bit_rate"`
	} `json:"format"`
	Streams []struct {
		CodecType  string `json:"codec_type"`
		CodecName  string `json:"codec_name"`
		Channels   int    `json:"channels"`
		SampleRate string `json:"sample_rate"`
	} `json:"streams"`
}
