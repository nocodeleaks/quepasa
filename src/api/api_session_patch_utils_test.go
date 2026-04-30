package api

import (
	"testing"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

func TestBuildSessionConfigurationPatchFromCreateRequest(t *testing.T) {
	groups := whatsapp.TrueBooleanType
	devel := true

	patch := buildSessionConfigurationPatch(&InfoCreateRequest{
		Groups: &groups,
		Devel:  &devel,
	})

	if patch == nil {
		t.Fatalf("expected patch")
	}

	if patch.Groups == nil || *patch.Groups != groups {
		t.Fatalf("expected groups to be mapped")
	}

	if patch.Devel == nil || *patch.Devel != devel {
		t.Fatalf("expected devel to be mapped")
	}
}

func TestBuildSessionConfigurationPatchFromPatchRequest(t *testing.T) {
	readReceipts := whatsapp.FalseBooleanType

	patch := buildSessionConfigurationPatch(&InfoPatchRequest{
		ReadReceipts: &readReceipts,
	})

	if patch == nil {
		t.Fatalf("expected patch")
	}

	if patch.ReadReceipts == nil || *patch.ReadReceipts != readReceipts {
		t.Fatalf("expected readreceipts to be mapped")
	}
}
