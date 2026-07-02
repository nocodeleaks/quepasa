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

func TestResolveGlobalTier(t *testing.T) {
	g := 3
	gt := "video"
	SetGlobalMessageConfigForTest(GlobalMessageConfig{StoreRetentionDays: &g, DispatchTypes: &gt})
	defer SetGlobalMessageConfigForTest(GlobalMessageConfig{}) // reset

	// global beats env
	r := ResolveMessageSettings(nil)
	if r.RetentionDays != 3 {
		t.Fatalf("global retention = %d, want 3", r.RetentionDays)
	}
	if !r.DispatchAllowed("video") || r.DispatchAllowed("text") {
		t.Fatal("global dispatch types not applied")
	}

	// per-caixa beats global
	s := &QpServer{}
	nine := int64(9)
	s.SetStoreRetentionDays(&nine)
	r = ResolveMessageSettings(&QpWhatsappServer{QpServer: s})
	if r.RetentionDays != 9 {
		t.Fatalf("per-caixa should beat global: got %d, want 9", r.RetentionDays)
	}
}
