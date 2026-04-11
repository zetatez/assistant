package network

import (
	"testing"
)

func TestGetSVRIP_WithValidNIC(t *testing.T) {
	ip := GetSVRIP("lo")
	if ip == "" {
		t.Error("expected non-empty IP for loopback")
	}
}

func TestGetSVRIP_WithInvalidNIC(t *testing.T) {
	ip := GetSVRIP("nonexistent_nic")
	if ip != DefaultSVRIP {
		t.Errorf("expected default IP %s, got %s", DefaultSVRIP, ip)
	}
}

func TestGetSVRIP_EmptyNIC(t *testing.T) {
	ip := GetSVRIP("")
	if ip != DefaultSVRIP {
		t.Errorf("expected default IP %s, got %s", DefaultSVRIP, ip)
	}
}
