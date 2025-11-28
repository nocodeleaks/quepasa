package library

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDuration_MarshalJSON(t *testing.T) {
	tests := []struct {
		name           string
		duration       Duration
		expectedSeconds int64
		expectedHuman  string
	}{
		{
			name:           "Less than 1 minute with milliseconds",
			duration:       Duration(54*time.Second + 636*time.Millisecond),
			expectedSeconds: 54,
			expectedHuman:  "55s", // Rounded to nearest second
		},
		{
			name:           "Exactly 1 minute",
			duration:       Duration(1 * time.Minute),
			expectedSeconds: 60,
			expectedHuman:  "1m0s",
		},
		{
			name:           "Hours, minutes and seconds",
			duration:       Duration(2*time.Hour + 30*time.Minute + 45*time.Second + 123*time.Millisecond),
			expectedSeconds: 9045,
			expectedHuman:  "2h30m45s",
		},
		{
			name:           "Multiple days",
			duration:       Duration(3*24*time.Hour + 5*time.Hour + 20*time.Minute),
			expectedSeconds: 278400,
			expectedHuman:  "77h20m0s",
		},
		{
			name:           "Zero duration",
			duration:       Duration(0),
			expectedSeconds: 0,
			expectedHuman:  "0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.duration)
			if err != nil {
				t.Fatalf("Failed to marshal duration: %v", err)
			}

			// Unmarshal to verify structure
			var result struct {
				Seconds int64  `json:"seconds"`
				Human   string `json:"human"`
			}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			// Verify seconds field
			if result.Seconds != tt.expectedSeconds {
				t.Errorf("Expected seconds=%d, got %d", tt.expectedSeconds, result.Seconds)
			}

			// Verify human field
			if result.Human != tt.expectedHuman {
				t.Errorf("Expected human=%q, got %q", tt.expectedHuman, result.Human)
			}

			t.Logf("Duration: %s -> {seconds: %d, human: %q}", 
				tt.duration.String(), result.Seconds, result.Human)
		})
	}
}

func TestDuration_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected Duration
	}{
		{
			name:     "Parse from seconds field",
			json:     `{"seconds": 54, "human": "54s"}`,
			expected: Duration(54 * time.Second),
		},
		{
			name:     "Parse larger duration",
			json:     `{"seconds": 9045, "human": "2h30m45s"}`,
			expected: Duration(9045 * time.Second),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Duration
			if err := json.Unmarshal([]byte(tt.json), &d); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if d != tt.expected {
				t.Errorf("Expected duration=%v, got %v", tt.expected, d)
			}
		})
	}
}
