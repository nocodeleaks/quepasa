package media

import (
	"bytes"
	"io"

	beep "github.com/gopxl/beep/v2"
	vorbis "github.com/gopxl/beep/v2/vorbis"
)

func GetStreamSeekFromVorbis(audioBytes []byte) (beep.Streamer, error) {
	f := bytes.NewReader(audioBytes)
	stringReadCloser := io.NopCloser(f)
	streamer, _, err := vorbis.Decode(stringReadCloser)
	if err != nil {
		return nil, err
	}
	defer streamer.Close()
	return streamer, nil
}
