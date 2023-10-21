package models

import (
	"errors"
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
		return nil, errors.New("User ID missing")
	}

	return WhatsappService.DB.Users.Find(user)
}

/*
<summary>

	Get a parameter from http.Request
	1º Url Param (/:parameter/)
	2º Url Query (?parameter=)
	3º Header (X-QUEPASA-PARAMETER)

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

//region TRIKCS

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

	URL has key, lowercase comparrison

</summary>
*/
func QueryHasKey(query *url.URL, key string) bool {
	for k := range query.Query() {
		if strings.ToLower(k) == strings.ToLower(key) {
			return true
		}
	}
	return false
}

/*
<summary>

	Get URL Value from Key, lowercase comparrison

</summary>
*/
func QueryGetValue(url *url.URL, key string) string {
	query := url.Query()
	for k := range query {
		if strings.ToLower(k) == strings.ToLower(key) {
			return query.Get(k)
		}
	}
	return ""
}

//endregion
