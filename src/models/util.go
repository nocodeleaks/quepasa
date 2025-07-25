package models

import (
	"strconv"
)

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

//endregion
