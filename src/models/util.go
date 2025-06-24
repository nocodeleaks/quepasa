package models

import (
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
)

// GetUser gets the user_id from the JWT and finds the
// corresponding user in the database
func GetFormUser(r *http.Request) (*QpUser, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return nil, err
	}

	user, ok := claims["user_id"].(string)
	if !ok {
		return nil, ErrFormUnauthenticated
	}

	return WhatsappService.DB.Users.Find(user)
}

/*
<summary>

	Get File Name From Http Request
	Getting from PATH => QUERY => HEADER

	* Trimmed & Url Unescaped

</summary>
*/
func GetFileName(r *http.Request) string {
	filename := GetRequestParameter(r, "filename")
	if len(filename) == 0 {
		mediatype := r.Header.Get("Content-Disposition")
		_, params, err := mime.ParseMediaType(mediatype)
		if err == nil {
			filename = params["filename"]
		}
	}

	if len(filename) > 0 {

		// url unescaping ...
		unescapedFilename, err := url.QueryUnescape(filename)
		if err == nil {
			filename = unescapedFilename
		}

		// trim white spaces from start and end
		filename = strings.TrimSpace(filename)
	}

	return filename
}

/*
<summary>

	Get a parameter from http.Request
	1º Url Param (/:parameter/)
	2º Url Query (?parameter=)
	3º Form
	4º Header (X-QUEPASA-PARAMETER)

</summary>
*/
func GetRequestParameter(r *http.Request, parameter string) string {
	// retrieve from url path parameter
	result := chi.URLParam(r, parameter)
	if len(result) == 0 {

		/// retrieve from url query parameter
		if QueryHasKey(r.URL, parameter) {
			result = QueryGetValue(r.URL, parameter)
		} else {

			if r.Form.Has(parameter) {
				result = r.Form.Get(parameter)
			} else {

				// retrieve from header parameter
				result = r.Header.Get("X-QUEPASA-" + strings.ToUpper(parameter))
			}
		}
	}

	// removing white spaces if exists
	return strings.TrimSpace(result)
}

// Getting ChatId from PATH => QUERY => HEADER
func GetChatId(r *http.Request) string {
	return GetRequestParameter(r, "chatid")
}

//region TRICKS

/*
<summary>

	Converts string to boolean with default value "false"

</summary>
*/
func ToBoolean(s string) bool {
	return ToBooleanWithDefault(s, false)
}

/*
<summary>

	Converts string to boolean with default value as argument

</summary>
*/
func ToBooleanWithDefault(s string, value bool) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return value
	}
	return b
}

/*
<summary>

	URL has key, lowercase comparison

</summary>
*/
func QueryHasKey(query *url.URL, key string) bool {
	for k := range query.Query() {
		if strings.EqualFold(k, key) {
			return true
		}
	}
	return false
}

/*
<summary>

	Get URL Value from Key, lowercase comparison
	* Url Unescaped

</summary>
*/
func QueryGetValue(rawUrl *url.URL, key string) string {
	query := rawUrl.Query()
	for k := range query {
		if strings.EqualFold(k, key) {
			value := query.Get(k)

			// unescape value
			unescapedValue, err := url.QueryUnescape(value)
			if err == nil {
				value = unescapedValue
			}

			return value
		}
	}
	return ""
}

//endregion
