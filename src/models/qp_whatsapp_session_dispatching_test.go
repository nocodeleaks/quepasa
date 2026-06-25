package models

import "testing"

func TestNewQpWhatsappSessionDispatchingFromDispatching_NilDispatchingReturnsNil(t *testing.T) {
	if got := NewQpWhatsappSessionDispatchingFromDispatching(nil, nil); got != nil {
		t.Fatal("expected nil wrapper for nil dispatching")
	}
}

func TestQpWhatsappSessionDispatching_SetSessionBindsLegacyServerField(t *testing.T) {
	dispatching := &QpDispatching{}
	session := &QpWhatsappSession{QpServer: &QpServer{Token: "sess-1"}}
	wrapper := NewQpWhatsappSessionDispatchingFromDispatching(dispatching, nil)
	if wrapper == nil {
		t.Fatal("expected non-nil wrapper")
	}

	wrapper.SetSession(session)

	legacy := (*QpWhatsappServerDispatching)(wrapper)
	if legacy.server != session {
		t.Fatal("expected session wrapper to bind the legacy server field")
	}
	if legacy.QpDispatching != dispatching {
		t.Fatal("expected session wrapper to preserve dispatching pointer")
	}
}
