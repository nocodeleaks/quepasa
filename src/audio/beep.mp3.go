package audio

import (
	"bytes"
	"io"

	beep "github.com/gopxl/beep/v2"
	mp3 "github.com/gopxl/beep/v2/mp3"

	newmp3 "github.com/tcolgate/mp3"
)

func GetStreamSeekFromMP3(audioBytes []byte) (beep.Streamer, error) {
	f := bytes.NewReader(audioBytes)
	stringReadCloser := io.NopCloser(f)
	streamer, _, err := mp3.Decode(stringReadCloser)
	if err != nil {
		return nil, err
	}
	defer streamer.Close()
	return streamer, nil
}

func GetDurationFromMP3(wavFileBytes []byte) (uint32, error) {
	reader := bytes.NewReader(wavFileBytes)
	dec := newmp3.NewDecoder(reader)
	var f newmp3.Frame
	skipped := 0

	var t float64
	for {

		if err := dec.Decode(&f, &skipped); err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}

		t = t + f.Duration().Seconds()
	}

	return uint32(t), nil
}
