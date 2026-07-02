package models

import "testing"

func TestDispatchDecision(t *testing.T) {
	all := ResolvedMessageSettings{} // store forever, all types
	if !dispatchDecision(all, "live", true, "text") {
		t.Fatal("realtime allowed type should dispatch")
	}
	if dispatchDecision(all, "history", false, "text") {
		t.Fatal("repeated history must not dispatch")
	}
	onlyText := ResolvedMessageSettings{DispatchTypes: map[string]bool{"text": true}}
	if dispatchDecision(onlyText, "live", true, "image") {
		t.Fatal("type not in filter must not dispatch")
	}
	if !dispatchDecision(onlyText, "live", true, "text") {
		t.Fatal("type in filter should dispatch")
	}
}
