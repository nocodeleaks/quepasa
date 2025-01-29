package whatsapp

import (
	"encoding/json"
	"fmt"
	"strings"
)

type WhatsappBoolean int

const (
	// False
	FalseBooleanType WhatsappBoolean = -1

	// Value not setted
	UnSetBooleanType WhatsappBoolean = 0

	// True
	TrueBooleanType WhatsappBoolean = 1
)

func (source WhatsappBoolean) MarshalJSON() ([]byte, error) {
	response := new(bool)
	switch source {
	case FalseBooleanType:
		*response = false
	case TrueBooleanType:
		*response = true
	default:
		response = nil
	}

	return json.Marshal(response)
}

// UnmarshalJSON parses fields that may be numbers or booleans.
func (source *WhatsappBoolean) UnmarshalJSON(b []byte) (err error) {

	json := string(b)
	json = strings.TrimSpace(json)
	json = strings.Trim(json, `"`)
	json = strings.ToLower(json)

	switch json {
	case "1", "t", "true", "yes":
		*source = TrueBooleanType
	case "-1", "f", "false", "no":
		*source = FalseBooleanType
	case "", "0":
		*source = UnSetBooleanType
	default:
		return fmt.Errorf("unknown boolean type: {%s}", json)
	}

	return
}

// converts to boolean, panic if invalid
func (source WhatsappBoolean) Boolean() bool {
	switch source {
	case FalseBooleanType:
		return false
	case TrueBooleanType:
		return true
	default:
		panic("invalid boolean value")
	}
}

// converts to boolean passing default value for unknown option
func (source WhatsappBoolean) ToBoolean(v bool) bool {
	switch source {
	case FalseBooleanType:
		return false
	case TrueBooleanType:
		return true
	default:
		return v
	}
}

func (source WhatsappBoolean) String() string {
	switch source {
	case FalseBooleanType:
		return "false"
	case TrueBooleanType:
		return "true"
	default:
		return "unset"
	}
}
