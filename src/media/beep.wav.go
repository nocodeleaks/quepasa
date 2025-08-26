package media

import (
	"bytes"
	"io"

	waveform "github.com/cettoana/go-waveform"
	"github.com/mattetti/audio/wav"

	beep "github.com/gopxl/beep/v2"
	beepwav "github.com/gopxl/beep/v2/wav"
)

func GetDurationFromWav(wavFileBytes []byte) (uint32, error) {
	f := bytes.NewReader(wavFileBytes)
	dec := wav.NewDecoder(f)
	time, err := dec.Duration()
	if err != nil {
		return 0, err
	}

	return uint32(time.Seconds()), nil
}

func IsValidWave(audioBytes []byte) bool {
	f := bytes.NewReader(audioBytes)
	dec := wav.NewDecoder(f)
	return dec.IsValidFile()
}

func GetWaveDetails(wavFileBytes []byte) AudioDetails {
	w := waveform.DecodeWav(wavFileBytes)

	return AudioDetails{
		Channels:      w.NumChannels,
		SampleRate:    w.SampleRate,
		BitsPerSample: w.BitsPerSample,
		Format:        w.WaveFormat.String(),
		DataChuckSize: w.DataChuckSize,
	}
}

func GetStreamSeekFromWav(audioBytes []byte) (beep.Streamer, error) {
	f := bytes.NewReader(audioBytes)
	stringReadCloser := io.NopCloser(f)
	streamer, _, err := beepwav.Decode(stringReadCloser)
	if err != nil {
		return nil, err
	}
	defer streamer.Close()
	return streamer, nil
}
