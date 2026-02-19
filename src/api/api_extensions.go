package api

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

type ReceiveMessageFilters struct {
	Exceptions  string
	Type        string
	Category    string
	Search      string
	ChatID      string
	MessageID   string
	TrackID     string
	FromMe      string
	FromHistory string
}

func GetReceiveMessageFilters(r *http.Request) ReceiveMessageFilters {
	return ReceiveMessageFilters{
		Exceptions:  strings.TrimSpace(library.GetRequestParameter(r, "exceptions")),
		Type:        strings.TrimSpace(library.GetRequestParameter(r, "type")),
		Category:    strings.TrimSpace(library.GetRequestParameter(r, "category")),
		Search:      strings.TrimSpace(library.GetRequestParameter(r, "search")),
		ChatID:      strings.TrimSpace(library.GetRequestParameter(r, "chatid")),
		MessageID:   strings.TrimSpace(library.GetRequestParameter(r, "messageid")),
		TrackID:     strings.TrimSpace(library.GetRequestParameter(r, "trackid")),
		FromMe:      strings.TrimSpace(library.GetRequestParameter(r, "fromme")),
		FromHistory: strings.TrimSpace(library.GetRequestParameter(r, "fromhistory")),
	}
}

func GetTimestamp(r *http.Request) (result int64, err error) {
	paramTimestamp := library.GetRequestParameter(r, "timestamp")
	if len(paramTimestamp) == 0 {
		paramLast := library.GetRequestParameter(r, "last")
		if len(paramLast) > 0 {
			last, err := strconv.ParseInt(paramLast, 10, 64)
			if err == nil {
				request := time.Now().UTC().Add(time.Duration(-last) * time.Minute)
				return request.Unix(), nil
			}
		}
	}

	result, err = StringToTimestamp(paramTimestamp)
	return

}

func StringToTimestamp(timestamp string) (result int64, err error) {
	if len(timestamp) > 0 {
		result, err = strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			if len(timestamp) > 0 {
				return
			} else {
				result = 0
			}
		}
	}
	return
}

/*
<summary>

	Retrieve messages with timestamp parameter
	Sorting then, by timestamp and id, desc

</summary>
*/
func GetOrderedMessages(server *models.QpWhatsappServer, timestamp int64) (messages []whatsapp.WhatsappMessage) {
	searchTime := time.Unix(timestamp, 0)
	messages = server.GetMessages(searchTime)
	sort.Sort(sort.Reverse(whatsapp.WhatsappOrderedMessages(messages)))
	return
}

/*
<summary>

	Retrieve messages with timestamp parameter and exceptions error filter
	Sorting then, by timestamp and id, desc
	ExceptionsFilter: "true" for messages with exceptions, "false" for messages without exceptions, "" for all messages

</summary>
*/
/*
<summary>

	Retrieve messages with timestamp parameter and exceptions error filter
	Sorting then, by timestamp and id, desc
	ExceptionsFilter: "true" for messages with exceptions, "false" for messages without exceptions, "" for all messages

</summary>
*/
func GetOrderedMessagesWithExceptionsFilter(server *models.QpWhatsappServer, timestamp int64, ExceptionsFilter string) (messages []whatsapp.WhatsappMessage) {
	filters := ReceiveMessageFilters{
		Exceptions: ExceptionsFilter,
	}
	return GetOrderedMessagesWithFilters(server, timestamp, filters)
}

func GetOrderedMessagesWithFilters(server *models.QpWhatsappServer, timestamp int64, filters ReceiveMessageFilters) (messages []whatsapp.WhatsappMessage) {
	searchTime := time.Unix(timestamp, 0)
	allMessages := server.GetMessages(searchTime)

	for _, msg := range allMessages {
		if !matchesMessage(msg, filters) {
			continue
		}
		messages = append(messages, msg)
	}

	sort.Sort(sort.Reverse(whatsapp.WhatsappOrderedMessages(messages)))
	return
}

func matchesMessage(msg whatsapp.WhatsappMessage, filters ReceiveMessageFilters) bool {
	if !matchesExceptionsFilter(msg, filters.Exceptions) {
		return false
	}

	if !matchesTypeFilter(msg, filters.Type) {
		return false
	}

	if !matchesCategoryFilter(msg, filters.Category) {
		return false
	}

	if !matchesBoolFilter(msg.FromMe, filters.FromMe) {
		return false
	}

	if !matchesBoolFilter(msg.FromHistory, filters.FromHistory) {
		return false
	}

	if !matchesContains(msg.Chat.Id, filters.ChatID) {
		return false
	}

	if !matchesContains(msg.Id, filters.MessageID) {
		return false
	}

	if !matchesContains(msg.TrackId, filters.TrackID) {
		return false
	}

	if !matchesSearch(msg, filters.Search) {
		return false
	}

	return true
}

func matchesExceptionsFilter(msg whatsapp.WhatsappMessage, filter string) bool {
	switch strings.ToLower(strings.TrimSpace(filter)) {
	case "true":
		return msg.HasExceptions()
	case "false":
		return !msg.HasExceptions()
	default:
		return true
	}
}

func matchesTypeFilter(msg whatsapp.WhatsappMessage, filter string) bool {
	normalized := normalizeFilterValue(filter)
	if len(normalized) == 0 {
		return true
	}

	msgType := normalizeFilterValue(msg.Type.String())
	for _, v := range splitFilterValues(normalized) {
		if v == msgType {
			return true
		}
	}
	return false
}

func matchesCategoryFilter(msg whatsapp.WhatsappMessage, filter string) bool {
	switch normalizeFilterValue(filter) {
	case "", "all":
		return true
	case "sent":
		return msg.FromMe
	case "received":
		return !msg.FromMe && msg.Type != whatsapp.SystemMessageType && msg.Type != whatsapp.UnhandledMessageType
	case "sync":
		return isSyncMessage(msg)
	case "unhandled":
		return msg.Type == whatsapp.UnhandledMessageType
	case "events":
		return msg.Type == whatsapp.SystemMessageType || msg.Type == whatsapp.UnhandledMessageType
	default:
		return true
	}
}

func isSyncMessage(msg whatsapp.WhatsappMessage) bool {
	if msg.Type != whatsapp.SystemMessageType {
		return false
	}

	text := strings.ToLower(msg.Text)
	return strings.Contains(text, "\"event\":\"sync_") || strings.Contains(text, "history synchronization event")
}

func matchesBoolFilter(value bool, filter string) bool {
	normalized := strings.ToLower(strings.TrimSpace(filter))
	if len(normalized) == 0 {
		return true
	}

	b, err := strconv.ParseBool(normalized)
	if err != nil {
		return true
	}
	return value == b
}

func matchesContains(value string, filter string) bool {
	filter = strings.TrimSpace(strings.ToLower(filter))
	if len(filter) == 0 {
		return true
	}
	return strings.Contains(strings.ToLower(value), filter)
}

func matchesSearch(msg whatsapp.WhatsappMessage, search string) bool {
	search = strings.TrimSpace(strings.ToLower(search))
	if len(search) == 0 {
		return true
	}

	if strings.Contains(strings.ToLower(msg.Id), search) {
		return true
	}
	if strings.Contains(strings.ToLower(msg.TrackId), search) {
		return true
	}
	if strings.Contains(strings.ToLower(msg.Text), search) {
		return true
	}
	if strings.Contains(strings.ToLower(msg.Chat.Id), search) {
		return true
	}
	if strings.Contains(strings.ToLower(msg.Chat.Title), search) {
		return true
	}
	if msg.Participant != nil {
		if strings.Contains(strings.ToLower(msg.Participant.Id), search) {
			return true
		}
		if strings.Contains(strings.ToLower(msg.Participant.Title), search) {
			return true
		}
	}
	for _, ex := range msg.Exceptions {
		if strings.Contains(strings.ToLower(ex), search) {
			return true
		}
	}
	if msg.Debug != nil {
		if strings.Contains(strings.ToLower(msg.Debug.Event), search) {
			return true
		}
		if strings.Contains(strings.ToLower(msg.Debug.Reason), search) {
			return true
		}
	}

	return false
}

func splitFilterValues(value string) []string {
	result := []string{}
	for _, item := range strings.Split(value, ",") {
		normalized := normalizeFilterValue(item)
		if len(normalized) == 0 {
			continue
		}
		result = append(result, normalized)
	}
	return result
}

func normalizeFilterValue(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	return normalized
}

/*
<summary>

	Find a system track identifier to follow the message
	Getting from PATH => QUERY => HEADER

</summary>
*/
func GetTrackId(r *http.Request) string {
	return library.GetRequestParameter(r, "trackid")
}

/*
<summary>

	Get Picture Identifier of contact
	Getting from PATH => QUERY => HEADER

</summary>
*/
func GetPictureId(r *http.Request) string {
	return library.GetRequestParameter(r, "pictureid")
}

/*
<summary>

	Get Token From Http Request
	Getting from PATH => QUERY => HEADER

</summary>
*/
func GetToken(r *http.Request) string {
	return library.GetRequestParameter(r, "token")
}

/*
<summary>

	Get Master Key From Http Request
	Getting from PATH => QUERY => HEADER

</summary>
*/
func GetMasterKey(r *http.Request) string {
	return library.GetRequestParameter(r, "masterkey")
}

/*
<summary>

	Get User From Http Request
	Getting from PATH => QUERY => HEADER
	If setted look after database, and throw errors
	If not setted does not throw errors and returns nil pointer

</summary>
*/
func GetUser(r *http.Request) (*models.QpUser, error) {
	username := library.GetRequestParameter(r, "user")
	if len(username) == 0 {
		ex := &BadRequestException{}
		ex.Prepend("missing user name parameter")
		return nil, ex
	}

	user, err := models.WhatsappService.DB.Users.Find(username)
	if err != nil {
		ex := &ApiExceptionBase{Inner: err}
		ex.Prependf("error for: %s", username)
		return nil, ex
	}

	return user, nil
}

/*
<summary>

	Get User From Http Request
	Getting from PATH => QUERY => HEADER
	If setted look after database, and throw errors
	If not setted does not throw errors and returns string empty

</summary>
*/
func GetUsername(r *http.Request) (string, ApiException) {
	user, err := GetUser(r)
	if err != nil {
		ex := &ApiExceptionBase{Inner: err}
		ex.Prepend("getting user name error")
		return "", ex
	}

	if user != nil {
		return user.Username, nil
	}

	ex := &ApiExceptionBase{}
	ex.Prependf("user not found: %s", user)
	return "", ex
}

/*
<summary>

	Get Message Id From Http Request
	Getting from PATH => QUERY => HEADER

</summary>
*/
func GetMessageId(r *http.Request) string {
	messageid := library.GetRequestParameter(r, "messageid")
	if len(messageid) == 0 {

		// compatibility with V3
		messageid = library.QueryGetValue(r.URL, "id")
	}
	return messageid
}

/*
<summary>

	Get Text Label From Http Request
	Getting from PATH => QUERY => FROM => HEADER

</summary>
*/
func GetTextParameter(r *http.Request) string {
	return library.GetRequestParameter(r, "text")
}

/*
<summary>

	Get In Reply From Http Request
	Getting from PATH => QUERY => FROM => HEADER

</summary>
*/
func GetInReplyParameter(r *http.Request) string {
	return library.GetRequestParameter(r, "inreply")
}

/*
<summary>

	Get a boolean indicating that cache should be used, From Http Request
	Getting from PATH => QUERY => FROM => HEADER

</summary>
*/
func GetCache(r *http.Request) bool {
	return models.ToBoolean(library.GetRequestParameter(r, "cache"))
}

/*
<summary>

	Get a boolean indicating that message id should be used as a prefix, defaults true
	Getting from PATH => QUERY => FROM => HEADER

</summary>
*/
func GetMessageIdAsPrefix(r *http.Request) bool {
	return models.ToBooleanWithDefault(library.GetRequestParameter(r, "messageidasprefix"), true)
}
