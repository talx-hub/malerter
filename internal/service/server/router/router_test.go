package router_test

import (
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

func newTestServer() *httptest.Server {
	log := logger.NewNopLogger()
	r := router.New(log, testSecret, constants.EmptyPath)
	r.SetRouter(testHandler{})
	return httptest.NewServer(r.GetRouter())
}

func TestRouter_HappyRoutes(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	tests := []struct {
		name      string
		method    string
		path      string
		wantCode  int
		wantXHead string
	}{
		{"GET /", http.MethodGet, "/", http.StatusTeapot, "GetAll"},
		{"GET /ping", http.MethodGet, "/ping", http.StatusTeapot, "Ping"},
		{"POST /value", http.MethodPost, "/value", http.StatusTeapot, "GetMetricJSON"},
		{"GET /value/gauge/ram", http.MethodGet, "/value/gauge/ram", http.StatusTeapot, "GetMetric"},
		{"POST /update", http.MethodPost, "/update", http.StatusTeapot, "DumpMetricJSON"},
		{"POST /update/gauge/ram/123", http.MethodPost, "/update/gauge/ram/123", http.StatusTeapot, "DumpMetric"},
		{"POST /updates", http.MethodPost, "/updates", http.StatusTeapot, "DumpMetricList"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, srv.URL+tt.path, http.NoBody)
			require.NoError(t, err)

			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/json")
				sig := signature.Hash([]byte(""), testSecret)
				req.Header.Set(constants.KeyHashSHA256, sig)
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
	srv := newTestServer()
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
