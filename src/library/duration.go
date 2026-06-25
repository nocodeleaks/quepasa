package library

import (
	"encoding/json"
	"time"
)

// Duration is a wrapper around time.Duration that serializes as an object with seconds and human-readable format
type Duration time.Duration

// MarshalJSON converts Duration to JSON object with "seconds" and "human" fields
func (d Duration) MarshalJSON() ([]byte, error) {
	duration := time.Duration(d)
	seconds := int64(duration.Seconds())

	// Format human-readable string based on duration
	var human string
	if seconds < 60 {
		// Less than 1 minute: show only seconds
		human = duration.Round(time.Second).String()
	} else if seconds < 3600 {
		// Less than 1 hour: show minutes and seconds
		human = duration.Round(time.Second).String()
	} else if seconds < 86400 {
		// Less than 1 day: show hours, minutes, seconds
		human = duration.Round(time.Second).String()
	} else {
		// 1 day or more: show days, hours, minutes
		human = duration.Round(time.Second).String()
	}

	return json.Marshal(struct {
		Seconds int64  `json:"seconds"`
		Human   string `json:"human"`
	}{
		Seconds: seconds,
		Human:   human,
	})
}

// UnmarshalJSON converts JSON back to Duration (from seconds field)
func (d *Duration) UnmarshalJSON(data []byte) error {
	var obj struct {
		Seconds int64  `json:"seconds"`
		Human   string `json:"human"`
	}

	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	*d = Duration(time.Duration(obj.Seconds) * time.Second)
	return nil
}

// Duration returns the underlying time.Duration value
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// Seconds returns the duration in seconds
func (d Duration) Seconds() float64 {
	return time.Duration(d).Seconds()
}

// String returns the human-readable representation
func (d Duration) String() string {
	return time.Duration(d).String()
}
