package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talx-hub/malerter/internal/backup"
	serverCfg "github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/repo"
	"net/http"
	"net/http/httptest"
	"testing"
)

func services(t *testing.T) (*httptest.Server, *backup.Backup) {
	cfg, ok := serverCfg.NewDirector().Build().(serverCfg.Builder)
	require.True(t, ok)

	zeroLogger, err := logger.New(cfg.LogLevel)
	require.NoError(t, err)

	rep := repo.NewMemRepository()

	bk, err := backup.New(cfg, rep)
	require.NoError(t, err)

	if cfg.Restore {
		bk.Restore()
	}
	ts := httptest.NewServer(metricRouter(rep, zeroLogger, bk))

	return ts, bk
}

func TestMetricRouter(t *testing.T) {
	ts, bk := services(t)
	defer func() {
		err := bk.Close()
		require.NoError(t, err)
	}()
	defer ts.Close()

	tests := []struct {
		method          string
		url             string
		body            string
		contentType     string
		statusWant      int
		encoding        string
		contentTypeWant string
	}{
		{
			method: http.MethodGet, url: "/",
			statusWant:      http.StatusOK,
			encoding:        "gzip",
			contentTypeWant: "text/html",
		},
		{
			method: http.MethodPost, url: "/",
			statusWant:      http.StatusMethodNotAllowed,
			encoding:        "",
			contentTypeWant: "",
		},
		{
			method: http.MethodPost, url: "/value/",
			contentType:     "application/json",
			body:            `{"id":"m1","type":"gauge","value":3.14}`,
			statusWant:      http.StatusNotFound,
			encoding:        "",
			contentTypeWant: "",
		},
		{
			method: http.MethodPost, url: "/value",
			contentType:     "application/json",
			body:            `{"id":"m1","type":"gauge","value":3.14}`,
			statusWant:      http.StatusNotFound,
			encoding:        "",
			contentTypeWant: "",
		},
		{
			method: http.MethodDelete, url: "/value",
			contentType:     "application/json",
			body:            `{"id":"m1","type":"gauge","value":3.14}`,
			statusWant:      http.StatusMethodNotAllowed,
			encoding:        "",
			contentTypeWant: "",
		},
		{
			method: http.MethodGet, url: "/value/",
			statusWant:      http.StatusMethodNotAllowed,
			encoding:        "",
			contentTypeWant: "",
		},
		{
			method: http.MethodGet, url: "/value",
			statusWant:      http.StatusMethodNotAllowed,
			encoding:        "",
			contentTypeWant: "",
		},
		{
			method: http.MethodGet, url: "/value/m1/gauge/",
			statusWant:      http.StatusNotFound,
			encoding:        "",
			contentTypeWant: "",
		},
		{
			method: http.MethodPost, url: "/value/m1/gauge",
			contentType:     "application/json",
			body:            `{"id":"m1","type":"gauge","value":3.14}`,
			statusWant:      http.StatusMethodNotAllowed,
			encoding:        "",
			contentTypeWant: "",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s %s", test.method, test.url), func(t *testing.T) {
			resp := testRequest(t, ts, test.method, test.url,
				test.body, test.contentType, test.encoding)
			assert.Equal(t, test.statusWant, resp.StatusCode)
			if test.statusWant != http.StatusOK {
				return
			}
			assert.Contains(
				t, resp.Header.Values("Content-Type"), test.contentTypeWant)
			if test.encoding != "" {
				assert.Contains(
					t, resp.Header.Values("Content-Encoding"), test.encoding)
			}
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server,
	method, url, body, contentType, encoding string) *http.Response {
	req, err := http.NewRequest(method, ts.URL+url, bytes.NewBufferString(body))
	require.NoError(t, err)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if encoding != "" {
		req.Header.Set("Accept-Encoding", encoding)
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()

	return resp
}
