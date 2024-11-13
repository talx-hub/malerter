package api

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repo"
	"github.com/talx-hub/malerter/internal/service"
	"github.com/talx-hub/malerter/internal/service/server"
)

func TestNewHTTPHandler(t *testing.T) {
	type args struct {
		service service.Service
	}
	tests := []struct {
		name string
		args args
		want *HTTPHandler
	}{
		{
			name: "simple constructor test #0",
			args: args{nil},
			want: &HTTPHandler{nil},
		},
		{
			name: "simple constructor test #1",
			args: args{service: server.NewMetricsDumper(nil)},
			want: &HTTPHandler{server.NewMetricsDumper(nil)},
		},
		{
			name: "simple constructor test #2",
			args: args{service: server.NewMetricsDumper(repo.NewMemRepository())},
			want: &HTTPHandler{server.NewMetricsDumper(repo.NewMemRepository())},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHTTPHandler(tt.args.service); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHTTPHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPHandler_DumpMetric(t *testing.T) {
	type test struct {
		url    string
		want   string
		status int
	}

	tests := []test{
		{"/update/counter/someMetric/123", "", 200},
		{"/update/gauge/someMetric/123.1", "", 200},
		{"/update/gauge/someMetric/1", "", 200},
		{"/update/gauge/1", "metric gauge/1/<nil> not found: metric name must be a string\n", 404},
		{"/update/WRONG/someMetric/1", "metric WRONG/someMetric/<nil> is incorrect: only counter and gauge types are allowed\n", 400},
		{"/update/counter/someMetric/1.0", "metric counter/someMetric/<nil> is incorrect: metric has invalid value\n", 400},
		{"/update/counter/someMetric", "metric counter/someMetric/<nil> not found: metric value is empty\n", 404},
		{"/update/counter/someMetric/9223372036854775808", "metric counter/someMetric/<nil> is incorrect: metric has invalid value\n", 400},
		{"/update/counter/someMetric/string", "metric counter/someMetric/<nil> is incorrect: invalid value <string>\n", 400},
	}
	rep := repo.NewMemRepository()
	serv := server.NewMetricsDumper(rep)
	handler := NewHTTPHandler(serv)
	ts := httptest.NewServer(http.HandlerFunc(handler.DumpMetric))
	defer ts.Close()

	wrongMethodTest := test{
		url:    "/update/gauge/someMetric/1",
		want:   "only POST requests are allowed\n",
		status: 400,
	}
	t.Run("wrong method test", func(t *testing.T) {
		resp, got := testRequest(t, ts, http.MethodGet, wrongMethodTest.url, "", nil)
		assert.Equal(t, wrongMethodTest.status, resp.StatusCode)
		assert.Equal(t, wrongMethodTest.want, got)
		if err := resp.Body.Close(); err != nil {
			log.Fatal(err)
		}
	})

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			resp, got := testRequest(t, ts, http.MethodPost, tt.url, "", nil)
			assert.Equal(t, tt.status, resp.StatusCode)
			assert.Equal(t, tt.want, got)
			if err := resp.Body.Close(); err != nil {
				log.Fatal(err)
			}
		})
	}
}

func TestHTTPHandler_GetMetric(t *testing.T) {
	type test struct {
		url    string
		want   string
		status int
	}

	tests := []test{
		{"/value/counter/mainQuestion", "42", 200},
		{"/value/gauge/pi", "3.14", 200},
		{"/value/wrong/pi", "metric wrong/pi/<nil> is incorrect: only counter and gauge types are allowed\n", 400},
		{"/value/gauge/wrong", "metric wrong(gauge): <nil> not found: \n", 404},
		{"/value/counter/wrong", "metric wrong(counter): <nil> not found: \n", 404},
		{"/value/counter", "metric /value/counter not found: incorrect URL\n", 404},
		{"/value/gauge", "metric /value/gauge not found: incorrect URL\n", 404},
	}
	m1, _ := model.NewMetric().FromValues("mainQuestion", model.MetricTypeCounter, int64(42))
	m2, _ := model.NewMetric().FromValues("pi", model.MetricTypeGauge, 3.14)
	repository := repo.NewMemRepository()
	_ = repository.Store(*m1)
	_ = repository.Store(*m2)

	dumper := server.NewMetricsDumper(repository)
	handler := NewHTTPHandler(dumper)
	ts := httptest.NewServer(http.HandlerFunc(handler.GetMetric))
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			resp, got := testRequest(t, ts, http.MethodGet, tt.url, "", nil)
			assert.Equal(t, tt.status, resp.StatusCode)
			assert.Equal(t, tt.want, got)
			if err := resp.Body.Close(); err != nil {
				log.Fatal(err)
			}
		})
	}
}

func TestHTTPHandler_DumpMetricJSON(t *testing.T) {
	tests := []struct {
		method       string
		url          string
		contentType  string
		body         string
		expectedCode int
		expectedBody string
	}{
		{
			method: http.MethodGet, url: "/update/",
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/update", contentType: "text/plain",
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/update", contentType: "application/json",
			body:         `{"id":"mainQuestion", "type":"counter", "delta":42}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"mainQuestion", "type":"counter", "delta":42}`,
		},
		{
			method: http.MethodPost, url: "/update", contentType: "application/json",
			body:         `{"id":"mainQuestion", "type":"counter", "delta":42}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"mainQuestion", "type":"counter", "delta":84}`,
		},
		{
			method: http.MethodPost, url: "/update", contentType: "application/json",
			body:         `{"type":"counter", "delta":42}`,
			expectedCode: http.StatusNotFound,
		},
		{
			method: http.MethodPost, url: "/update", contentType: "application/json",
			body:         `{"id":"42","type":"counter", "delta":42}`,
			expectedCode: http.StatusNotFound,
		},
		{
			method: http.MethodPost, url: "/update", contentType: "application/json",
			body:         `{"id":"m42","type":"wrong", "delta":42}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/update", contentType: "application/json",
			body:         `{"id":"m42","type":"counter", "delta":42.5}`,
			expectedCode: http.StatusInternalServerError,
		},
		{
			method: http.MethodPost, url: "/update", contentType: "application/json",
			body:         `{"id":"m42","type":"counter"}`,
			expectedCode: http.StatusNotFound,
		},
		{
			method: http.MethodPost, url: "/update", contentType: "application/json",
			body:         `{"id":"m42"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/update", contentType: "application/json",
			body:         `{"type":"counter"}`,
			expectedCode: http.StatusNotFound,
		},
		{
			method: http.MethodPost, url: "/update", contentType: "application/json",
			body:         `{"delta":42.5}`,
			expectedCode: http.StatusInternalServerError,
		},
		{
			method: http.MethodPost, url: "/update", contentType: "application/json",
			body:         ``,
			expectedCode: http.StatusInternalServerError,
		},
		{
			method: http.MethodPost, url: "/update", contentType: "application/json",
			body:         `{"id":"m42","type":"counter", "delta":42, "value":3.14}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/update", contentType: "application/json",
			body:         `{"id":"m42","type":"wrong", "delta":"42"}`,
			expectedCode: http.StatusInternalServerError,
		},
		{
			method: http.MethodPost, url: "/update", contentType: "application/json",
			body:         `{"id":"pi", "type":"gauge", "value":3.14}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"pi", "type":"gauge", "value":3.14}`,
		},
		{
			method: http.MethodPost, url: "/update", contentType: "application/json",
			body:         `{"id":"pi", "type":"gauge", "value":3.1415926}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"pi", "type":"gauge", "value":3.1415926}`,
		},
		{
			method: http.MethodPost, url: "/update", contentType: "application/json",
			body:         `{"id":"pi", "type":"gauge", "delta":3}`,
			expectedCode: http.StatusBadRequest,
		},
	}

	repository := repo.NewMemRepository()
	dumper := server.NewMetricsDumper(repository)
	handler := NewHTTPHandler(dumper)
	testServer := httptest.NewServer(http.HandlerFunc(handler.DumpMetricJSON))
	defer testServer.Close()

	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			resp, got := testRequest(t, testServer, test.method, test.url, test.contentType, &test.body)
			assert.Equal(t, test.expectedCode, resp.StatusCode)
			if test.expectedCode == http.StatusOK {
				assert.JSONEq(t, test.expectedBody, got)
			}
			if err := resp.Body.Close(); err != nil {
				log.Fatal(err)
			}
		})
	}
}

func TestHTTPHandler_GetMetricJSON(t *testing.T) {
	tests := []struct {
		method       string
		url          string
		contentType  string
		body         string
		expectedCode int
		expectedBody string
	}{
		{
			method: http.MethodGet, url: "/value",
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/value", contentType: "text/plain",
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/value", contentType: "application/json",
			body:         `{"id":"m42", "type":"counter"}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"m42", "type":"counter", "delta":42}`,
		},
		{
			method: http.MethodPost, url: "/value", contentType: "application/json",
			body:         `{"type":"counter"}`,
			expectedCode: http.StatusNotFound,
		},
		{
			method: http.MethodPost, url: "/value", contentType: "application/json",
			body:         `{"id":"42","type":"counter"}`,
			expectedCode: http.StatusNotFound,
		},
		{
			method: http.MethodPost, url: "/value", contentType: "application/json",
			body:         `{"id":"m42","type":"wrong"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/value", contentType: "application/json",
			body:         `{"id":"m42"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/value", contentType: "application/json",
			body:         `{"id":"wrong","type":"counter"}`,
			expectedCode: http.StatusNotFound,
		},
		{
			method: http.MethodPost, url: "/value", contentType: "application/json",
			body:         `{"id":"pi", "type":"gauge"}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"pi", "type":"gauge", "value":3.14}`,
		},
	}

	repository := repo.NewMemRepository()
	m1, _ := model.NewMetric().FromValues("m42", model.MetricTypeCounter, int64(42))
	m2, _ := model.NewMetric().FromValues("pi", model.MetricTypeGauge, 3.14)
	_ = repository.Store(*m1)
	_ = repository.Store(*m2)

	dumper := server.NewMetricsDumper(repository)
	handler := NewHTTPHandler(dumper)
	testServer := httptest.NewServer(http.HandlerFunc(handler.GetMetricJSON))
	defer testServer.Close()

	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			resp, got := testRequest(t, testServer, test.method, test.url, test.contentType, &test.body)
			assert.Equal(t, test.expectedCode, resp.StatusCode)
			if test.expectedCode == http.StatusOK {
				assert.JSONEq(t, test.expectedBody, got)
			}
			if err := resp.Body.Close(); err != nil {
				log.Fatal(err)
			}
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server,
	method, path, contentType string, body *string) (*http.Response, string) {
	var request *http.Request
	var err error
	if body != nil {
		request, err = http.NewRequest(method, ts.URL+path, bytes.NewBuffer([]byte(*body)))
	} else {
		request, err = http.NewRequest(method, ts.URL+path, nil)
	}
	require.NoError(t, err)
	request.Header.Set("Content-Type", contentType)

	resp, err := ts.Client().Do(request)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestHTTPHandler_GetAll(t *testing.T) {
	tests := []struct {
		method         string
		url            string
		body           string
		expectedStatus int
		expectedBody   string
	}{
		{
			method: http.MethodPost, url: "/",
			expectedStatus: http.StatusBadRequest,
		},
		{
			method: http.MethodGet, url: "/",
			expectedStatus: http.StatusOK,
			body:           "<html>\n\t<body>\n\t\t<p>m42(counter): 42</p>\n\t</body>\n</html>",
			expectedBody:   "<html>\n\t<body>\n\t\t<p>m42(counter): 42</p>\n\t</body>\n</html>",
		},
	}

	repository := repo.NewMemRepository()
	m1, _ := model.NewMetric().FromValues("m42", model.MetricTypeCounter, int64(42))
	_ = repository.Store(*m1)

	dumper := server.NewMetricsDumper(repository)
	handler := NewHTTPHandler(dumper)
	testServer := httptest.NewServer(http.HandlerFunc(handler.GetAll))

	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			resp, got := testRequest(t, testServer, test.method, test.url, "", &test.body)
			assert.Equal(t, test.expectedStatus, resp.StatusCode)
			if test.expectedStatus == http.StatusOK {
				assert.Equal(t, test.expectedBody, got)
			}
			if err := resp.Body.Close(); err != nil {
				log.Fatal(err)
			}
		})
	}
}
