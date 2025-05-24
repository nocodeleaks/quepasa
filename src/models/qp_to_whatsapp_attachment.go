package models

import (
	"fmt"
	"path/filepath"
	"strings"

	audio "github.com/nocodeleaks/quepasa/audio"
	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

type QpToWhatsappAttachment struct {
	Attach *whatsapp.WhatsappAttachment
	Debug  []string `json:"debug,omitempty"`
}

func (source *QpToWhatsappAttachment) AttachSecureAndCustomize() {
	attach := source.Attach
	if attach == nil {
		source.Debug = append(source.Debug, "[warn][AttachSecureAndCustomize] nil attach")
		return
	}

	source.Debug = append(source.Debug, fmt.Sprintf("[trace][AttachSecureAndCustomize] initial mime type: %s, filename: %s", attach.Mimetype, attach.FileName))

	var contentMime string
	content := attach.GetContent()
	if content != nil {
		contentMime = library.GetMimeTypeFromContent(*content)
		source.Debug = append(source.Debug, fmt.Sprintf("[debug][AttachSecureAndCustomize] detected mime type from content: %s", contentMime))
	}

	var requestExtension string
	if len(attach.FileName) > 0 {
		fileNameNormalized := strings.ToLower(attach.FileName) // important, because some bitches do capitalize filenames
		requestExtension = filepath.Ext(fileNameNormalized)
		source.Debug = append(source.Debug, fmt.Sprintf("[debug][AttachSecureAndCustomize] detected extension: %s, from filename: %s", requestExtension, attach.FileName))
	} else if len(attach.Mimetype) > 0 {
		requestExtension, _ = library.TryGetExtensionFromMimeType(attach.Mimetype)
		source.Debug = append(source.Debug, fmt.Sprintf("[debug][AttachSecureAndCustomize] detected extension from request mime type: %s", requestExtension))
	} else if len(contentMime) > 0 {
		requestExtension, _ = library.TryGetExtensionFromMimeType(contentMime)
		source.Debug = append(source.Debug, fmt.Sprintf("[debug][AttachSecureAndCustomize] detected extension from content mime type: %s", requestExtension))
	}

	if len(contentMime) > 0 {

		// downloaded pdf, issue by @Marcelo
		if attach.Mimetype == "application/x-www-form-urlencoded" && contentMime == "application/pdf" {
			attach.Mimetype = contentMime
			source.Debug = append(source.Debug, fmt.Sprintf("[info][AttachSecureAndCustomize] updating downloaded mime type from content: %s", contentMime))
		}

		if strings.HasPrefix(contentMime, "text/xml") && requestExtension == ".svg" {
			contentMime = "image/svg+xml"
		}

		if len(attach.Mimetype) == 0 {
			attach.Mimetype = contentMime
			source.Debug = append(source.Debug, fmt.Sprintf("[debug][AttachSecureAndCustomize] updating empty mime type from content: %s", contentMime))
		}

		contentExtension, success := library.TryGetExtensionFromMimeType(contentMime)
		if success {
			source.Debug = append(source.Debug, fmt.Sprintf("[debug][AttachSecureAndCustomize] content extension: %s", contentExtension))

			// if was passed a filename without extension
			if len(requestExtension) == 0 && len(attach.FileName) > 0 {
				source.Debug = append(source.Debug, fmt.Sprintf("[info][AttachSecureAndCustomize] missing extension for attachment (%s), using from content: %s :: content mime: %s", attach.FileName, contentExtension, contentMime))

				attach.Mimetype = contentMime
				attach.FileName += contentExtension
			} else {
				// validating mime information
				if !IsValidExtensionFor(requestExtension, contentExtension) {
					// invalid attachment
					source.Debug = append(source.Debug, fmt.Sprintf("[warn][AttachSecureAndCustomize] invalid extension for attachment, request extension: %s (%s) != content extension: %s :: content mime: %s, revalidating for security", requestExtension, attach.FileName, contentExtension, contentMime))
					attach.Mimetype = contentMime
					attach.FileName = whatsapp.InvalidFilePrefix + library.GenerateFileNameFromMimeType(contentMime)
				}
			}
		}
	}

	// setting a filename if not found before
	if len(attach.FileName) == 0 {
		attach.FileName = library.GenerateFileNameFromMimeType(attach.Mimetype)
		source.Debug = append(source.Debug, fmt.Sprintf("[debug][AttachSecureAndCustomize] empty file name, generating a new one based on mime type: %s, file name: %s", attach.Mimetype, attach.FileName))
	}

	// if pdf mime contains extra info
	if strings.HasPrefix(attach.Mimetype, "application/pdf;") {
		source.Debug = append(source.Debug, fmt.Sprintf("[info][AttachSecureAndCustomize] removing extra information from pdf mime type: %s", attach.Mimetype))
		attach.Mimetype = strings.Split(attach.Mimetype, ";")[0]
	}

	source.Debug = append(source.Debug, fmt.Sprintf("[debug][AttachSecureAndCustomize] resolved mime type: %s, filename: %s", attach.Mimetype, attach.FileName))
}

// Audio Formatting ...
func (source *QpToWhatsappAttachment) AttachAudioTreatmentTesting() {

	if audio.AreAudioToolsAvailable() {

		audioInfo, err := audio.GetAudioInfoFromBytes(*source.Attach.GetContent())
		if err != nil {
			log.Errorf("Erro ao obter as informações de áudio a partir dos bytes: %v", err)
			return
		}

		seconds := audioInfo.Duration.Seconds()
		if seconds > 0 {
			if source.Attach.Seconds == 0 {
				source.Attach.Seconds = uint32(seconds)
			}
		}

		source.Attach.WaveForm, err = audio.GenerateWaveform(*source.Attach.GetContent())
		if err != nil {
			log.Errorf("error generating waveform from bytes: %v", err)
			return
		}

		log.Tracef("\n--- Informações de Áudio ---\n")
		log.Tracef("Duração:    %s\n", audioInfo.Duration)
		log.Tracef("MIME Type:  %s\n", audioInfo.MIMEType)
		log.Tracef("Canais:     %d\n", audioInfo.Channels)
		log.Tracef("Sample Rate: %d Hz\n", audioInfo.SampleRate)
	}
}

func (source *QpToWhatsappAttachment) AttachAudioTreatment() {
	attach := source.Attach
	if attach == nil {
		source.Debug = append(source.Debug, "[warn][AttachAudioTreatment] nil attach")
		return
	}

	if audio.IsAudioMIMEType(source.Attach.Mimetype) {
		source.AttachAudioTreatmentTesting()
	}

	if IsCompatibleWithPTT(attach.Mimetype) {
		source.Debug = append(source.Debug, fmt.Sprintf("[trace][AttachAudioTreatment] mime type is compatible for ptt: %s", attach.Mimetype))

		// set compatible audios to be sent as ptt
		ForceCompatiblePTT := ENV.UseCompatibleMIMEsAsAudio()
		if ForceCompatiblePTT && !attach.IsValidAudio() {
			source.Debug = append(source.Debug, fmt.Sprintf("[info][AttachAudioTreatment] setting that it should be sent as ptt, regards its incompatible mime type: %s", attach.Mimetype))
			attach.SetPTTCompatible(true)
		}
	}

	if IsCompatibleWithPTT(attach.Mimetype) || attach.IsValidAudio() {
		if ENV.Testing() {
			source.AudioDetails()
		}
	}
}

func (source *QpToWhatsappAttachment) AudioDetails() {
	debug := audio.GetAudioDetails(source.Attach)
	source.Debug = append(source.Debug, debug...)
}
