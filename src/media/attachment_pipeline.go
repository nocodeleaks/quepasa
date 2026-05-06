package media

import (
	"fmt"
	"path/filepath"
	"strings"

	environment "github.com/nocodeleaks/quepasa/environment"
	library "github.com/nocodeleaks/quepasa/library"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

type QpToWhatsappAttachment struct {
	Attach *whatsapp.WhatsappAttachment
	Debug  []string `json:"debug,omitempty"`
}

func IsValidExtensionFor(request string, content string) bool {
	switch {
	case
		request == ".csv" && content == ".txt",
		request == ".jpg" && content == ".jpeg",
		request == ".jpeg" && content == ".jpg",
		request == ".json" && content == ".txt",
		request == ".oga" && content == ".webm",
		request == ".oga" && content == ".ogx",
		request == ".opus" && content == ".ogx",
		request == ".ovpn" && content == ".txt",
		request == ".pdf" && content == ".txt",
		request == ".sql" && content == ".txt",
		request == ".svg" && content == ".xml",
		request == ".xml" && content == ".txt":
		return true
	}

	return request == content
}

func IsCompatibleWithPTT(mime string) bool {
	mimeOnly := strings.Split(mime, ";")[0]

	for _, item := range whatsapp.WhatsappMIMEAudioPTTCompatible {
		if item == mimeOnly {
			return true
		}
	}

	return false
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
		fileNameNormalized := strings.ToLower(attach.FileName)
		requestExtension = filepath.Ext(fileNameNormalized)
		source.Debug = append(source.Debug, fmt.Sprintf("[debug][AttachSecureAndCustomize] detected extension: %s, from filename: %s", requestExtension, attach.FileName))
	} else if len(attach.Mimetype) > 0 {
		// Strip MIME parameters (e.g. "; codecs=opus") before extension lookup so that
		// "audio/ogg; codecs=opus" resolves to the same extension as plain "audio/ogg".
		baseMime := strings.TrimSpace(strings.SplitN(attach.Mimetype, ";", 2)[0])
		requestExtension, _ = library.TryGetExtensionFromMimeType(baseMime)
		source.Debug = append(source.Debug, fmt.Sprintf("[debug][AttachSecureAndCustomize] detected extension from request mime type: %s (base: %s)", requestExtension, baseMime))
	} else if len(contentMime) > 0 {
		requestExtension, _ = library.TryGetExtensionFromMimeType(contentMime)
		source.Debug = append(source.Debug, fmt.Sprintf("[debug][AttachSecureAndCustomize] detected extension from content mime type: %s", requestExtension))
	}

	if len(contentMime) > 0 {
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

			if len(requestExtension) == 0 && len(attach.FileName) > 0 {
				source.Debug = append(source.Debug, fmt.Sprintf("[info][AttachSecureAndCustomize] missing extension for attachment (%s), using from content: %s :: content mime: %s", attach.FileName, contentExtension, contentMime))

				attach.Mimetype = contentMime
				attach.FileName += contentExtension
			} else {
				if !IsValidExtensionFor(requestExtension, contentExtension) {
					source.Debug = append(source.Debug, fmt.Sprintf("[warn][AttachSecureAndCustomize] invalid extension for attachment, request extension: %s (%s) != content extension: %s :: content mime: %s, revalidating for security", requestExtension, attach.FileName, contentExtension, contentMime))
					attach.Mimetype = contentMime
					attach.FileName = whatsapp.InvalidFilePrefix + library.GenerateFileNameFromMimeType(contentMime)
				}
			}
		}
	}

	if len(attach.FileName) == 0 {
		attach.FileName = library.GenerateFileNameFromMimeType(attach.Mimetype)
		source.Debug = append(source.Debug, fmt.Sprintf("[debug][AttachSecureAndCustomize] empty file name, generating a new one based on mime type: %s, file name: %s", attach.Mimetype, attach.FileName))
	}

	if strings.HasPrefix(attach.Mimetype, "application/pdf;") {
		source.Debug = append(source.Debug, fmt.Sprintf("[info][AttachSecureAndCustomize] removing extra information from pdf mime type: %s", attach.Mimetype))
		attach.Mimetype = strings.Split(attach.Mimetype, ";")[0]
	}

	source.Debug = append(source.Debug, fmt.Sprintf("[debug][AttachSecureAndCustomize] resolved mime type: %s, filename: %s", attach.Mimetype, attach.FileName))
}

func (source *QpToWhatsappAttachment) AttachAudioTreatmentTesting() {
	if AreAudioToolsAvailable() {
		audioInfo, err := GetAudioInfoFromBytes(*source.Attach.GetContent())
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

		// Only overwrite WaveForm if generation succeeds; preserve any client-provided waveform on error.
		if wf, wfErr := GenerateWaveform(*source.Attach.GetContent()); wfErr != nil {
			log.Errorf("error generating waveform from bytes: %v", wfErr)
		} else if len(wf) > 0 {
			source.Attach.WaveForm = wf
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

	// If the MIME is already the canonical PTT format, mark as PTT immediately so
	// GetMessageType returns AudioMessageType regardless of further processing.
	if attach.IsValidPTT() {
		attach.SetPTTCompatible(true)
		source.Debug = append(source.Debug, fmt.Sprintf("[info][AttachAudioTreatment] canonical PTT MIME detected (%s), marking as PTT", attach.Mimetype))
	}

	forceAudioAsPTT := environment.Settings.General.ForceAudioAsPTT

	if forceAudioAsPTT && ShouldConvertToPTT(attach.Mimetype) {
		source.Debug = append(source.Debug, fmt.Sprintf("[info][AttachAudioTreatment] FORCE_AUDIO_AS_PTT enabled, converting %s to OGG Opus via ffmpeg", attach.Mimetype))

		content := attach.GetContent()
		if content != nil && len(*content) > 0 {
			originalMime := attach.Mimetype
			convertedData, err := ConvertToOggOpus(*content)
			if err != nil {
				source.Debug = append(source.Debug, fmt.Sprintf("[error][AttachAudioTreatment] failed to convert %s to OGG Opus: %v", originalMime, err))
				log.Errorf("Failed to convert %s to OGG Opus: %v", originalMime, err)
			} else {
				originalSize := len(*content)
				attach.SetContent(&convertedData)
				attach.Mimetype = whatsapp.WhatsappPTTMime
				attach.FileLength = uint64(len(convertedData))

				if len(attach.FileName) > 0 {
					ext := filepath.Ext(attach.FileName)
					if len(ext) > 0 {
						attach.FileName = attach.FileName[:len(attach.FileName)-len(ext)] + ".ogg"
					}
				}

				attach.SetPTTCompatible(true)

				source.Debug = append(source.Debug, fmt.Sprintf("[success][AttachAudioTreatment] converted to OGG Opus PTT. Original: %d bytes (%s), New: %d bytes (%s)", originalSize, originalMime, len(convertedData), whatsapp.WhatsappPTTMime))
			}
		} else {
			source.Debug = append(source.Debug, "[warn][AttachAudioTreatment] no content available for audio conversion")
		}
	}

	if IsAudioMIMEType(source.Attach.Mimetype) {
		source.AttachAudioTreatmentTesting()
	}

	if IsCompatibleWithPTT(attach.Mimetype) {
		source.Debug = append(source.Debug, fmt.Sprintf("[trace][AttachAudioTreatment] mime type is compatible for ptt: %s", attach.Mimetype))

		forceCompatiblePTT := environment.Settings.General.UseCompatibleMIMEsAsAudio()
		if forceCompatiblePTT && !attach.IsValidAudio() {
			source.Debug = append(source.Debug, fmt.Sprintf("[info][AttachAudioTreatment] setting that it should be sent as ptt, regards its incompatible mime type: %s", attach.Mimetype))
			attach.SetPTTCompatible(true)
		}
	}

	if IsCompatibleWithPTT(attach.Mimetype) || attach.IsValidAudio() {
		if environment.Settings.General.Testing {
			source.AudioDetails()
		}
	}
}

func (source *QpToWhatsappAttachment) AudioDetails() {
	debug := GetAudioDetails(source.Attach)
	source.Debug = append(source.Debug, debug...)
}

func (source *QpToWhatsappAttachment) AttachImageTreatment() {
	attach := source.Attach
	if attach == nil {
		source.Debug = append(source.Debug, "[warn][AttachImageTreatment] nil attach")
		return
	}

	if !environment.Settings.General.ConvertPNGToJPG {
		source.Debug = append(source.Debug, "[trace][AttachImageTreatment] PNG to JPG conversion is disabled in settings, returning without image validation")
		return
	}

	if !ShouldConvertImage(attach.Mimetype, attach.FileName) {
		source.Debug = append(source.Debug, fmt.Sprintf("[trace][AttachImageTreatment] PNG image conversion not required, current mime: %s, filename: %s", attach.Mimetype, attach.FileName))
		return
	}

	source.Debug = append(source.Debug, fmt.Sprintf("[info][AttachImageTreatment] PNG image detected, attempting conversion to JPG. Current mime: %s, filename: %s", attach.Mimetype, attach.FileName))

	content := attach.GetContent()
	if content == nil || len(*content) == 0 {
		source.Debug = append(source.Debug, "[warn][AttachImageTreatment] no content available for PNG conversion")
		return
	}

	jpgData, newMime, err := ConvertPngToJpg(*content)
	if err != nil {
		source.Debug = append(source.Debug, fmt.Sprintf("[error][AttachImageTreatment] failed to convert PNG to JPG: %v", err))
		log.Errorf("Failed to convert PNG to JPG: %v", err)
		return
	}

	originalSize := len(*content)
	newSize := len(jpgData)

	attach.SetContent(&jpgData)
	attach.Mimetype = newMime
	attach.FileLength = uint64(newSize)

	if len(attach.FileName) > 0 {
		lowerFileName := strings.ToLower(attach.FileName)
		if strings.HasSuffix(lowerFileName, ".png") {
			baseFileName := attach.FileName[:len(attach.FileName)-4]
			attach.FileName = baseFileName + ".jpg"
		}
	}

	source.Debug = append(source.Debug, fmt.Sprintf("[success][AttachImageTreatment] PNG successfully converted to JPG. Original size: %d bytes, new size: %d bytes, new mime: %s, new filename: %s", originalSize, newSize, newMime, attach.FileName))
}
