package api

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	library "github.com/nocodeleaks/quepasa/library"
	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

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

	Retrieve messages with timestamp parameter and dispatch error filter
	Sorting then, by timestamp and id, desc
	dispatchErrorFilter: "true" for messages with dispatch errors, "false" for messages without dispatch errors, "" for all messages

</summary>
*/
/*
<summary>

	Retrieve messages with timestamp parameter and dispatch error filter
	Sorting then, by timestamp and id, desc
	dispatchErrorFilter: "true" for messages with dispatch errors, "false" for messages without dispatch errors, "" for all messages

</summary>
*/
func GetOrderedMessagesWithDispatchFilter(server *models.QpWhatsappServer, timestamp int64, dispatchErrorFilter string) (messages []whatsapp.WhatsappMessage) {
	searchTime := time.Unix(timestamp, 0)
	allMessages := server.GetMessages(searchTime)

	// Filter messages based on dispatch error status
	switch dispatchErrorFilter {
	case "true":
		// Return only messages with dispatch errors
		for _, msg := range allMessages {
			if msg.HasDispatchError() {
				messages = append(messages, msg)
			}
		}
	case "false":
		// Return only messages without dispatch errors
		for _, msg := range allMessages {
			if !msg.HasDispatchError() {
				messages = append(messages, msg)
			}
		}
	default:
		// Return all messages (no filter)
		messages = allMessages
	}

	sort.Sort(sort.Reverse(whatsapp.WhatsappOrderedMessages(messages)))
	return
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
