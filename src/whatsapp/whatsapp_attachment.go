package whatsapp

import (
	"strings"
)

type WhatsappAttachment struct {
	content *[]byte `json:"-"`

	// means that it can be downloaded from whatsapp servers
	CanDownload bool `json:"-"`

	Mimetype string `json:"mime"`

	// important to navigate throw content, declared file length
	FileLength uint64 `json:"filelength"`

	// document
	FileName string `json:"filename,omitempty"`

	// video | image | location (base64 image)
	JpegThumbnail string `json:"thumbnail,omitempty"`

	// audio/video
	Seconds uint32 `json:"seconds,omitempty"`

	// audio, used for define that this attach should be sent as ptt compatible, regards its incompatible mime type
	ptt bool `json:"-"`

	// location msgs
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Sequence  int64   `json:"sequence,omitempty"` // live location

	// Public access url helper content
	Url string `json:"url,omitempty"`

	WaveForm []byte `json:"waveform,omitempty"`
}

func (source *WhatsappAttachment) GetContent() *[]byte {
	return source.content
}

func (source *WhatsappAttachment) SetContent(content *[]byte) {
	source.content = content
}

func (source *WhatsappAttachment) HasContent() bool {
	return nil != source.content
}

// used at receive.tmpl view
func (source *WhatsappAttachment) IsValidSize() bool {

	var length int
	if source.FileLength > 0 {
		length = int(source.FileLength)
	} else if source.content != nil {
		length = len(*source.content)
	}

	if length > 500 {
		return true
	}

	// there are many simple vcards with low bytes
	// vcard | plain | etc
	if strings.HasPrefix(source.Mimetype, "text/") && length > 50 {
		return true
	}

	return false
}

func (source *WhatsappAttachment) SetPTTCompatible(value bool) {
	source.ptt = value
}

func (source *WhatsappAttachment) IsPTTCompatible() bool {
	return source.ptt
}

func (source *WhatsappAttachment) IsValidAudio() bool {
	if source.IsValidPTT() {
		return true
	}

	// switch for basic mime type, ignoring suffix
	mimeOnly := strings.Split(source.Mimetype, ";")[0]

	for _, item := range WhatsappMIMEAudio {
		if item == mimeOnly {
			return true
		}
	}

	return false
}

func (source *WhatsappAttachment) IsValidPTT() bool {
	return source.Mimetype == WhatsappPTTMime
}
