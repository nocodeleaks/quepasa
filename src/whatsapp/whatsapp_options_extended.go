package whatsapp

import "time"

// Whatsapp service options, setted on start, so if want to changed then, you have to restart the entire service
type WhatsappOptionsExtended struct {

	// should handle groups messages
	Groups WhatsappBooleanExtended `json:"groups,omitempty"`

	// should handle broadcast messages
	Broadcasts WhatsappBooleanExtended `json:"broadcasts,omitempty"`

	// should emit read receipts
	ReadReceipts WhatsappBooleanExtended `json:"readreceipts,omitempty"`

	// should handle calls
	Calls WhatsappBooleanExtended `json:"calls,omitempty"`

	// should send markread requests
	ReadUpdate bool `json:"readupdate"`

	// nil for no sync, 0 for all, X for specific days
	HistorySync *uint32 `json:"historysync,omitempty"`

	// default log level
	LogLevel string `json:"loglevel,omitempty"`

	// default presence status
	Presence string `json:"presence,omitempty"`

	// should dispatch unhandled messages
	DispatchUnhandled bool `json:"dispatchunhandled,omitempty"`
}

func (source WhatsappOptionsExtended) IsDefault() bool {
	return source.Groups.Equals(UnSetBooleanType) &&
		source.Broadcasts.Equals(UnSetBooleanType) &&
		source.ReadReceipts.Equals(UnSetBooleanType) &&
		source.Calls.Equals(UnSetBooleanType) &&
		!source.ReadUpdate &&
		source.HistorySync == nil &&
		!source.DispatchUnhandled &&
		len(source.LogLevel) == 0
}

/*
<summary>
	default options from environment variables
	should be set on main
</summary>
*/
var Options WhatsappOptionsExtended

func (source WhatsappOptionsExtended) HandleCalls(local WhatsappBoolean) bool {
	switch source.Calls {
	case ForcedFalseBooleanType:
		return false
	case ForcedTrueBooleanType:
		return true
	default:
		if local != UnSetBooleanType {
			return local.Boolean()
		}

		return source.Calls.ToBoolean(WhatsappCalls)
	}
}

func (source WhatsappOptionsExtended) HandleReadReceipts(local WhatsappBoolean) bool {
	switch source.ReadReceipts {
	case ForcedFalseBooleanType:
		return false
	case ForcedTrueBooleanType:
		return true
	default:
		if local != UnSetBooleanType {
			return local.Boolean()
		}

		return source.ReadReceipts.ToBoolean(WhatsappReadReceipts)
	}
}

func (source WhatsappOptionsExtended) HandleGroups(local WhatsappBoolean) bool {
	switch source.Groups {
	case ForcedFalseBooleanType:
		return false
	case ForcedTrueBooleanType:
		return true
	default:
		if local != UnSetBooleanType {
			return local.Boolean()
		}

		return source.Groups.ToBoolean(WhatsappGroups)
	}
}

func (source WhatsappOptionsExtended) HandleBroadcasts(local WhatsappBoolean) bool {
	switch source.Broadcasts {
	case ForcedFalseBooleanType:
		return false
	case ForcedTrueBooleanType:
		return true
	default:
		if local != UnSetBooleanType {
			return local.Boolean()
		}

		return source.Broadcasts.ToBoolean(WhatsappBroadcasts)
	}
}

func (source WhatsappOptionsExtended) HandleHistory(mts uint64) bool {
	if source.HistorySync != nil {
		days := *source.HistorySync
		if days == 0 {
			return true
		}

		current := time.Now()
		limit := current.AddDate(0, 0, -int(days))
		if int64(mts) > limit.Unix() {
			return true
		}
	}

	return false
}
