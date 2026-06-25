package whatsapp

type WhatsappBooleanExtended int

const (
	// Forced False
	ForcedFalseBooleanType WhatsappBooleanExtended = -2

	// Forced True
	ForcedTrueBooleanType WhatsappBooleanExtended = 2
)

// converts to boolean passing default value for unknown option
func (source WhatsappBooleanExtended) ToBoolean(v bool) bool {
	switch source {
	case ForcedFalseBooleanType:
		return false
	case ForcedTrueBooleanType:
		return true
	default:
		return WhatsappBoolean(source).ToBoolean(v)
	}
}

func (source WhatsappBooleanExtended) String() string {
	switch source {
	case ForcedFalseBooleanType:
		return "forcedfalse"
	case ForcedTrueBooleanType:
		return "forcedtrue"
	default:
		return WhatsappBoolean(source).String()
	}
}

func (source WhatsappBooleanExtended) Equals(v WhatsappBoolean) bool {
	switch source {
	case ForcedFalseBooleanType, ForcedTrueBooleanType:
		return false
	default:
		return WhatsappBoolean(source) == v
	}
}

func (source WhatsappBooleanExtended) Compare(item WhatsappBoolean, value bool) bool {
	switch source {
	case ForcedFalseBooleanType:
		return false
	case ForcedTrueBooleanType:
		return true
	default:
		return item.ToBoolean(value)
	}
}
