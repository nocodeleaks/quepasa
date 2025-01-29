package models

type QpSendResponseV2 struct {
	QpResponse
	ID   string       `json:"message_id"`
	Date int          `json:"date,omitempty"`
	From QPEndpointV2 `json:"from,omitempty"`
	Chat QPEndpointV2 `json:"chat,omitempty"`

	// Para compatibilidade apenas
	PreviusV1 QPSendResult `json:"result,omitempty"`
}
