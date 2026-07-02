package models

import "testing"

func TestResolvedStoreAndDispatch(t *testing.T) {
	r := ResolvedMessageSettings{RetentionDays: -1}
	if r.Store() {
		t.Fatal("none must not store")
	}
	r = ResolvedMessageSettings{RetentionDays: 0}
	if !r.Store() || !r.ExpiryFor().IsZero() {
		t.Fatal("forever must store with zero (never-expire) time")
	}
	r = ResolvedMessageSettings{RetentionDays: 7}
	if !r.Store() || r.ExpiryFor().IsZero() {
		t.Fatal("N days must store with a non-zero expiry")
	}
	r = ResolvedMessageSettings{DispatchTypes: map[string]bool{"text": true}}
	if !r.DispatchAllowed("text") || r.DispatchAllowed("image") {
		t.Fatal("only listed types dispatch")
	}
	r = ResolvedMessageSettings{}
	if !r.DispatchAllowed("anything") {
		t.Fatal("empty dispatch types = all allowed")
	}
}
