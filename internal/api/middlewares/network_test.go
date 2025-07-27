package middlewares

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/talx-hub/malerter/internal/logger"
)

func createIPNet(cidr string) *net.IPNet {
	_, ipnet, _ := net.ParseCIDR(cidr)
	return ipnet
}

func TestCheckNetwork_IPAllowed(t *testing.T) {
	ipNet := createIPNet("192.168.1.0/24")

	middleware := CheckNetwork(ipNet, logger.NewNopLogger())
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("X-Real-IP", "192.168.1.42")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "ok") {
		t.Errorf("unexpected body: %s", rr.Body.String())
	}
}

func TestCheckNetwork_IPDenied(t *testing.T) {
	ipNet := createIPNet("192.168.1.0/24")

	middleware := CheckNetwork(ipNet, logger.NewNopLogger())
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("X-Real-IP", "10.0.0.1")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}
}

func TestCheckNetwork_NoIPProvided(t *testing.T) {
	ipNet := createIPNet("192.168.1.0/24")

	middleware := CheckNetwork(ipNet, logger.NewNopLogger())
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}
}

func TestCheckNetwork_NilIPNet(t *testing.T) {
	middleware := CheckNetwork(nil, logger.NewNopLogger())
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("open"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "open") {
		t.Errorf("unexpected body: %s", rr.Body.String())
	}
}
