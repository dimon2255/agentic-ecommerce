package admin

import (
	"net/http/httptest"
	"testing"
)

func TestRealIP_XForwardedFor_TrustsLastEntry(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "spoofed-by-attacker, 10.0.0.1, 203.0.113.50")

	got := realIP(req)
	if got != "203.0.113.50" {
		t.Errorf("expected last entry 203.0.113.50, got %s", got)
	}
}

func TestRealIP_XForwardedFor_SingleEntry(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50")

	got := realIP(req)
	if got != "203.0.113.50" {
		t.Errorf("expected 203.0.113.50, got %s", got)
	}
}

func TestRealIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "198.51.100.10")

	got := realIP(req)
	if got != "198.51.100.10" {
		t.Errorf("expected 198.51.100.10, got %s", got)
	}
}

func TestRealIP_FallbackToRemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:54321"

	got := realIP(req)
	if got != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1 (port stripped), got %s", got)
	}
}

func TestRealIP_RemoteAddrNoPort(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1"

	got := realIP(req)
	if got != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1, got %s", got)
	}
}

func TestRealIP_XForwardedFor_PrioritizedOverXRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50")
	req.Header.Set("X-Real-IP", "198.51.100.10")

	got := realIP(req)
	if got != "203.0.113.50" {
		t.Errorf("expected X-Forwarded-For value 203.0.113.50, got %s", got)
	}
}
