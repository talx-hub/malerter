package router_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/service/server/router"
	"github.com/talx-hub/malerter/pkg/signature"
)

const testSecret = "test-secret"
const testTrustedSubnet = "127.0.0.0/24"

type stubHandler struct {
	name string
}

func (s stubHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("X-Handler", s.name)
	w.WriteHeader(http.StatusTeapot)
}

type testHandler struct{}

func (testHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	stubHandler{"GetAll"}.ServeHTTP(w, r)
}
func (testHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	stubHandler{"GetMetric"}.ServeHTTP(w, r)
}
func (testHandler) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	stubHandler{"GetMetricJSON"}.ServeHTTP(w, r)
}
func (testHandler) DumpMetric(w http.ResponseWriter, r *http.Request) {
	stubHandler{"DumpMetric"}.ServeHTTP(w, r)
}
func (testHandler) DumpMetricJSON(w http.ResponseWriter, r *http.Request) {
	stubHandler{"DumpMetricJSON"}.ServeHTTP(w, r)
}
func (testHandler) DumpMetricList(w http.ResponseWriter, r *http.Request) {
	stubHandler{"DumpMetricList"}.ServeHTTP(w, r)
}
func (testHandler) Ping(w http.ResponseWriter, r *http.Request) { stubHandler{"Ping"}.ServeHTTP(w, r) }

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	_, subnet, err := net.ParseCIDR(testTrustedSubnet)
	require.NoError(t, err)

	r := router.New(logger.NewNopLogger(), subnet, testSecret, constants.EmptyPath)
	r.SetRouter(testHandler{})
	return httptest.NewServer(r.GetRouter())
}

func TestRouter_HappyRoutes(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	tests := []struct {
		name           string
		method         string
		path           string
		fromNotTrusted bool
		wantCode       int
		wantXHead      string
	}{
		{"GET /", http.MethodGet, "/", false, http.StatusTeapot, "GetAll"},
		{"GET /ping", http.MethodGet, "/ping", false, http.StatusTeapot, "Ping"},
		{"POST /value", http.MethodPost, "/value", false, http.StatusTeapot, "GetMetricJSON"},
		{"GET /value/gauge/ram", http.MethodGet, "/value/gauge/ram", false, http.StatusTeapot, "GetMetric"},
		{"POST /update", http.MethodPost, "/update", false, http.StatusTeapot, "DumpMetricJSON"},
		{"POST /update", http.MethodPost, "/update", true, http.StatusForbidden, ""},
		{"POST /update/gauge/ram/123", http.MethodPost, "/update/gauge/ram/123", false, http.StatusTeapot, "DumpMetric"},
		{"POST /updates", http.MethodPost, "/updates", false, http.StatusTeapot, "DumpMetricList"},
		{"POST /updates", http.MethodPost, "/updates", true, http.StatusForbidden, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, srv.URL+tt.path, http.NoBody)
			require.NoError(t, err)

			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/json")
				sig := signature.Hash([]byte(""), testSecret)
				req.Header.Set(constants.KeyHashSHA256, sig)
				if !tt.fromNotTrusted {
					req.Header.Set("X-Real-IP", "127.0.0.2")
				}
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()

			assert.Equal(t, tt.wantCode, resp.StatusCode)
			assert.Equal(t, tt.wantXHead, resp.Header.Get("X-Handler"))
		})
	}
}

func TestRouter_not_check_network(t *testing.T) {
	r := router.New(logger.NewNopLogger(), nil, testSecret, constants.EmptyPath)
	r.SetRouter(testHandler{})
	srv := httptest.NewServer(r.GetRouter())
	defer srv.Close()

	tests := []struct {
		name           string
		method         string
		path           string
		fromNotTrusted bool
		wantCode       int
		wantXHead      string
	}{
		{"GET /", http.MethodGet, "/", false, http.StatusTeapot, "GetAll"},
		{"GET /ping", http.MethodGet, "/ping", false, http.StatusTeapot, "Ping"},
		{"POST /value", http.MethodPost, "/value", false, http.StatusTeapot, "GetMetricJSON"},
		{"GET /value/gauge/ram", http.MethodGet, "/value/gauge/ram", false, http.StatusTeapot, "GetMetric"},
		{"POST /update", http.MethodPost, "/update", false, http.StatusTeapot, "DumpMetricJSON"},
		{"POST /update", http.MethodPost, "/update", true, http.StatusTeapot, "DumpMetricJSON"},
		{"POST /update/gauge/ram/123", http.MethodPost, "/update/gauge/ram/123", false, http.StatusTeapot, "DumpMetric"},
		{"POST /updates", http.MethodPost, "/updates", false, http.StatusTeapot, "DumpMetricList"},
		{"POST /updates", http.MethodPost, "/updates", true, http.StatusTeapot, "DumpMetricList"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, srv.URL+tt.path, http.NoBody)
			require.NoError(t, err)

			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/json")
				sig := signature.Hash([]byte(""), testSecret)
				req.Header.Set(constants.KeyHashSHA256, sig)
				if tt.fromNotTrusted {
					req.Header.Set("X-Real-IP", "1.0.0.2")
				} else {
					req.Header.Set("X-Real-IP", "127.0.0.2")
				}
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()

			assert.Equal(t, tt.wantCode, resp.StatusCode)
			assert.Equal(t, tt.wantXHead, resp.Header.Get("X-Handler"))
		})
	}
}

func TestRouter_WrongRoutes(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	tests := []struct {
		name     string
		method   string
		path     string
		wantCode int
	}{
		{"NotFound route", http.MethodGet, "/not-exist", http.StatusNotFound},
		{"NotFound on incomplete path", http.MethodGet, "/value/", http.StatusMethodNotAllowed},
		{"MethodNotAllowed", http.MethodPut, "/", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, srv.URL+tt.path, http.NoBody)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()

			assert.Equal(t, tt.wantCode, resp.StatusCode)
		})
	}
}
