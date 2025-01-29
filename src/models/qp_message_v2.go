package models

// Mensagem no formato QuePasa
// Utilizada na API do QuePasa para troca com outros sistemas
type QpMessageV2 struct {
	ID        string `json:"message_id"`
	Timestamp uint64 `json:"timestamp"`

	// Whatsapp que gerencia a bagaça toda
	Controller QPEndpointV2 `json:"controller"`

	// Endereço garantido que deve receber uma resposta
	ReplyTo QPEndpointV2 `json:"replyto"`

	// Se a msg foi postado em algum grupo ? quem postou !
	Participant QPEndpointV2 `json:"participant,omitempty"`

	// Fui eu quem enviou a msg ?
	FromMe bool `json:"fromme"`

	// Texto da msg
	Text string `json:"text"`

	Attachment *QPAttachmentV1 `json:"attachment,omitempty"`

	Chat QPChatV2 `json:"chat"`
}

type ByTimestampV2 []QpMessageV2

func (m ByTimestampV2) Len() int           { return len(m) }
func (m ByTimestampV2) Less(i, j int) bool { return m[i].Timestamp > m[j].Timestamp }
func (m ByTimestampV2) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
