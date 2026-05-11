package models

import "testing"

func TestSessionIntentNoneIsNotStop(t *testing.T) {
	if SessionIntentNone.IsStopRequested() {
		t.Fatal("SessionIntentNone should not be stop-requested")
	}
}

func TestSessionIntentNoneIsNotDelete(t *testing.T) {
	if SessionIntentNone.IsDeleteRequested() {
		t.Fatal("SessionIntentNone should not be delete-requested")
	}
}

func TestSessionIntentStopIsStop(t *testing.T) {
	if !SessionIntentStop.IsStopRequested() {
		t.Fatal("SessionIntentStop should be stop-requested")
	}
}

func TestSessionIntentStopIsNotDelete(t *testing.T) {
	if SessionIntentStop.IsDeleteRequested() {
		t.Fatal("SessionIntentStop should not be delete-requested")
	}
}

func TestSessionIntentDeleteIsStop(t *testing.T) {
	if !SessionIntentDelete.IsStopRequested() {
		t.Fatal("SessionIntentDelete should also be stop-requested")
	}
}

func TestSessionIntentDeleteIsDelete(t *testing.T) {
	if !SessionIntentDelete.IsDeleteRequested() {
		t.Fatal("SessionIntentDelete should be delete-requested")
	}
}

func TestSessionIntentString(t *testing.T) {
	cases := []struct {
		intent   SessionIntent
		expected string
	}{
		{SessionIntentNone, "None"},
		{SessionIntentStop, "Stop"},
		{SessionIntentDelete, "Delete"},
		{SessionIntent(99), "Unknown"},
	}

	for _, tc := range cases {
		if got := tc.intent.String(); got != tc.expected {
			t.Errorf("SessionIntent(%d).String() = %q, want %q", tc.intent, got, tc.expected)
		}
	}
}
