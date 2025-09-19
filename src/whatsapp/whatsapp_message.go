package whatsapp

import (
	"strings"
	"time"
)

// Mensagem no formato QuePasa
// Utilizada na API do QuePasa para troca com outros sistemas
type WhatsappMessage struct {

	// original message from source service
	Content any `json:"-"`

	InfoForHistory any `json:"-"`

	Id      string `json:"id"`                // Upper text msg id
	TrackId string `json:"trackid,omitempty"` // Optional id of the system that send that message

	Timestamp time.Time           `json:"timestamp"`
	Type      WhatsappMessageType `json:"type"`

	// Em qual chat (grupo ou direct) essa msg foi postada, para onde devemos responder
	Chat WhatsappChat `json:"chat"`

	// If this message was posted on a Group, Who posted it !
	Participant *WhatsappChat `json:"participant,omitempty"`

	// Message text if exists
	Text string `json:"text,omitempty"`

	Attachment *WhatsappAttachment `json:"attachment,omitempty"`

	// Do i send that ?
	// From any connected device and api
	FromMe bool `json:"fromme"`

	// Sended via api
	FromInternal bool `json:"frominternal"`

	// Generated from history sync
	FromHistory bool `json:"fromhistory,omitempty"`

	// Edited message
	Edited bool `json:"edited,omitempty"`

	// How many times this message was forwarded
	ForwardingScore uint32 `json:"forwardingscore,omitempty"`

	// Msg in reply of another ? Message ID
	InReply string `json:"inreply,omitempty"`

	// Msg in reply preview
	Synopsis string `json:"synopsis,omitempty"`

	// Delivered, Read, Imported statuses
	Status WhatsappMessageStatus `json:"status,omitempty"`

	// Url if exists
	Url *WhatsappMessageUrl `json:"url,omitempty"`

	Ads *WhatsappMessageAds `json:"ads,omitempty"`

	// WhatsApp ID of the sender
	Wid string `json:"wid,omitempty"`

	// Extra information for custom messages
	Info any `json:"info,omitempty"`

	Poll *WhatsappPoll `json:"poll,omitempty"` // Poll if exists

	// Debug information for debug events
	Debug *WhatsappMessageDebug `json:"debug,omitempty"`

	Exceptions []string `json:"exceptions,omitempty"`
}

//region ORDER BY TIMESTAMP

// Ordering by (Timestamp) and then (Id)
type WhatsappOrderedMessages []WhatsappMessage

func (m WhatsappOrderedMessages) Len() int { return len(m) }
func (m WhatsappOrderedMessages) Less(i, j int) bool {
	if m[i].Timestamp.Equal(m[j].Timestamp) {
		return m[i].Id < m[j].Id
	}
	return m[i].Timestamp.Before(m[j].Timestamp)
}
func (m WhatsappOrderedMessages) Swap(i, j int) { m[i], m[j] = m[j], m[i] }

//endregion

//region IMPLEMENT WHATSAPP SEND RESPONSE INTERFACE

func (source WhatsappMessage) GetId() string { return source.Id }

// Get the time of server processed message
func (source WhatsappMessage) GetTime() time.Time { return source.Timestamp }

// Get the time on unix timestamp format
func (source WhatsappMessage) GetTimestamp() uint64 { return uint64(source.Timestamp.Unix()) }

//endregion

func (source *WhatsappMessage) GetChatId() string {
	return source.Chat.Id
}

func (source *WhatsappMessage) GetParticipantId() string {
	if source.Participant == nil {
		return ""
	}
	return source.Participant.Id
}

func (source *WhatsappMessage) GetText() string {
	return source.Text
}

// Indicates if the message has any status information
// *Trick to help in Views
func (source *WhatsappMessage) HasStatus() bool {
	return source != nil && len(source.Status) > 0
}

// Indicates if the message has url information
// *Trick to help in Views
func (source *WhatsappMessage) HasUrl() bool {
	return source != nil && source.Url != nil && len(source.Url.Reference) > 0
}

func (source *WhatsappMessage) HasAttachment() bool {
	// this attachment is a pointer to correct show info on deserialized
	attach := source.Attachment
	return attach != nil && len(attach.Mimetype) > 0
}

func (source *WhatsappMessage) GetSource() any {
	return source.Content
}

func (source *WhatsappMessage) FromGroup() bool {
	return strings.HasSuffix(source.Chat.Id, WHATSAPP_SERVERDOMAIN_GROUP_SUFFIX)
}

func (source *WhatsappMessage) FromAds() bool {
	return source.Ads != nil
}

func (source *WhatsappMessage) FromBroadcast() bool {
	if source.Chat.Id == "status" {
		return true
	}

	if source.Chat.Id == "status@broadcast" {
		return true
	}

	if strings.HasSuffix(source.Chat.Id, "@newsletter") {
		return true
	}

	return false
}

//endregion

//region DISPATCH ERROR MANAGEMENT

// MarkExceptions marks the message as having a dispatch error
func (source *WhatsappMessage) MarkExceptions() {
	source.Exceptions = append(source.Exceptions, "Dispatch error occurred")
}

// MarkExceptionsWithMessage marks the message as having a dispatch error with a specific message
func (source *WhatsappMessage) MarkExceptionsWithMessage(message string) {
	source.Exceptions = append(source.Exceptions, message)
}

// ClearExceptions clears all dispatch errors
func (source *WhatsappMessage) ClearExceptions() {
	source.Exceptions = nil
}

// HasExceptions checks if the message has any dispatch errors
func (source *WhatsappMessage) HasExceptions() bool {
	return len(source.Exceptions) > 0
}

// GetExceptions returns all dispatch error messages
func (source *WhatsappMessage) GetExceptions() []string {
	return source.Exceptions
}

//endregion

func (source *WhatsappMessage) GetAttachment() *WhatsappAttachment {
	return source.Attachment
}
