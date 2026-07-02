package environment

import "testing"

func TestParseStoreRetention(t *testing.T) {
	cases := map[string]int{"": 0, "0": 0, "forever": 0, "none": -1, "-1": -1, "10": 10, "garbage": 0}
	for in, want := range cases {
		if got := parseStoreRetention(in); got != want {
			t.Errorf("parseStoreRetention(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestParseDispatchTypes(t *testing.T) {
	got := parseDispatchTypes(" text, Image ,audio ")
	if len(got) != 3 || !got["text"] || !got["image"] || !got["audio"] {
		t.Fatalf("parseDispatchTypes = %+v, want text/image/audio (lowercased, trimmed)", got)
	}
	if parseDispatchTypes("") != nil {
		t.Fatalf("empty DISPATCH_TYPES should be nil (all allowed)")
	}
}
