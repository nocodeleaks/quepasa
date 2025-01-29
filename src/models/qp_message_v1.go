package models

// Mensagem no formato QuePasa
// Utilizada na API do QuePasa para troca com outros sistemas
type QPMessageV1 struct {
	ID        string `json:"message_id"`
	Timestamp uint64 `json:"timestamp"`

	// Whatsapp que gerencia a bagaça toda
	Controller QPEndpointV1 `json:"controller"`

	// Endereço garantido que deve receber uma resposta
	ReplyTo QPEndpointV1 `json:"replyto"`

	// Se a msg foi postado em algum grupo ? quem postou !
	Participant QPEndpointV1 `json:"participant,omitempty"`

	// Fui eu quem enviou a msg ?
	FromMe bool `json:"fromme"`

	// Texto da msg
	Text string `json:"text"`

	Attachment *QPAttachmentV1 `json:"attachment,omitempty"`
}

type ByTimestampV1 []QPMessageV1

func (m ByTimestampV1) Len() int           { return len(m) }
func (m ByTimestampV1) Less(i, j int) bool { return m[i].Timestamp > m[j].Timestamp }
func (m ByTimestampV1) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

func (source *QPMessageV1) ToV2() *QpMessageV2 {
	message := &QpMessageV2{}
	message.ID = source.ID
	message.Timestamp = source.Timestamp
	message.Controller = source.Controller.GetQPEndPointV2()
	message.ReplyTo = source.ReplyTo.GetQPEndPointV2()
	message.Participant = source.Participant.GetQPEndPointV2()
	message.FromMe = source.FromMe
	message.Text = source.Text
	message.Attachment = source.Attachment
	message.Chat = source.ReplyTo.ToQPChatV2()

	return message
}

//region IMPLEMENT INTERFACE WHATSAPP MESSAGE

func (source *QPMessageV1) GetText() string {
	return source.Text
}

func (source *QPMessageV1) GetChatID() string {
	return source.ReplyTo.ID
}

// Check if that msg has a valid attachment
func (source *QPMessageV1) HasAttachment() bool {
	// this attachment is a pointer to correct show info on deserialized
	attach := source.Attachment
	return attach != nil && attach.Length > 0
}

//endregion
