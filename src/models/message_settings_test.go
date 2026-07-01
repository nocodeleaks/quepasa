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

// serverForResolveTest wraps a QpServer into a *QpWhatsappServer; QpWhatsappServer
// embeds *QpServer, so the config fields are reachable via server.QpServer.
func serverForResolveTest(s *QpServer) *QpWhatsappServer {
	return &QpWhatsappServer{QpServer: s}
}

func TestResolvePerServerOverride(t *testing.T) {
	s := &QpServer{}
	five := int64(5)
	s.SetStoreRetentionDays(&five)
	types := "audio"
	s.SetDispatchTypes(&types)
	sv := serverForResolveTest(s)
	r := ResolveMessageSettings(sv)
	if r.RetentionDays != 5 {
		t.Fatalf("override retention = %d, want 5", r.RetentionDays)
	}
	if !r.DispatchAllowed("audio") || r.DispatchAllowed("text") {
		t.Fatalf("override dispatch types not applied: %+v", r.DispatchTypes)
	}
	_ = ResolveMessageSettings(nil) // nil must not panic
}
