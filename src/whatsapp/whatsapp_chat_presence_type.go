package whatsapp

import "encoding/json"

type WhatsappChatPresenceType uint

const (
	WhatsappChatPresenceTypePaused WhatsappChatPresenceType = iota
	WhatsappChatPresenceTypeText
	WhatsappChatPresenceTypeAudio
)

func (Type WhatsappChatPresenceType) String() string {
	switch Type {
	case WhatsappChatPresenceTypeText:
		return "text"
	case WhatsappChatPresenceTypeAudio:
		return "audio"
	default:
		return "paused"
	}
}

func (s *WhatsappChatPresenceType) Parse(str string) {
	switch str {
	case "text":
		*s = WhatsappChatPresenceTypeText
	case "audio":
		*s = WhatsappChatPresenceTypeAudio
	default:
		*s = WhatsappChatPresenceTypePaused
	}
}

func (s WhatsappChatPresenceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *WhatsappChatPresenceType) UnmarshalJSON(data []byte) error {
	var number uint
	var str string
	if err := json.Unmarshal(data, &number); err != nil {
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		} else {
			s.Parse(str)
		}
	} else {
		*s = WhatsappChatPresenceType(number)
	}
	return nil
}
