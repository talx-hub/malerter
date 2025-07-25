package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/internal/api/handlers"
	serverCfg "github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/internal/service/server/router"
)

func setupServices(t *testing.T) *httptest.Server {
	t.Helper()

	cfg, ok := serverCfg.NewDirector().Build().(serverCfg.Builder)
	require.True(t, ok)

	zeroLogger, err := logger.New(cfg.LogLevel)
	require.NoError(t, err)

	rep := memory.New(zeroLogger, nil)
	require.NotNil(t, rep)

	chiRouter := router.New(zeroLogger, constants.NoSecret, cfg.CryptoKeyPath)
	chiRouter.SetRouter(handlers.NewHTTPHandler(rep, zeroLogger))
	ts := httptest.NewServer(chiRouter.GetRouter())

	return ts
}

func TestMetricRouter(t *testing.T) {
	ts := setupServices(t)
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
			encoding:        constants.EncodingGzip,
			contentTypeWant: constants.ContentTypeHTML,
		},
		{
			method: http.MethodPost, url: "/",
			statusWant:      http.StatusMethodNotAllowed,
			encoding:        "",
			contentTypeWant: "",
		},
		{
			method: http.MethodPost, url: "/value/",
			contentType:     constants.ContentTypeJSON,
			body:            `{"id":"m1","type":"gauge","value":3.14}`,
			statusWant:      http.StatusNotFound,
			encoding:        "",
			contentTypeWant: "",
		},
		{
			method: http.MethodPost, url: "/value",
			contentType:     constants.ContentTypeJSON,
			body:            `{"id":"m1","type":"gauge","value":3.14}`,
			statusWant:      http.StatusNotFound,
			encoding:        "",
			contentTypeWant: "",
		},
		{
			method: http.MethodDelete, url: "/value",
			contentType:     constants.ContentTypeJSON,
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
			contentType:     constants.ContentTypeJSON,
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
			defer func() {
				err := resp.Body.Close()
				require.NoError(t, err)
			}()

			assert.Equal(t, test.statusWant, resp.StatusCode)
			if test.statusWant != http.StatusOK {
				return
			}
			assert.Contains(
				t, resp.Header.Values(constants.KeyContentType), test.contentTypeWant)
			if test.encoding != "" {
				assert.Contains(
					t, resp.Header.Values(constants.KeyContentEncoding), test.encoding)
			}
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server,
	method, url, body, contentType, encoding string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, ts.URL+url, bytes.NewBufferString(body))
	require.NoError(t, err)
	if contentType != "" {
		req.Header.Set(constants.KeyContentType, contentType)
	}
	if encoding != "" {
		req.Header.Set(constants.KeyAcceptEncoding, encoding)
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	return resp
}
