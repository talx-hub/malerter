package api

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/memory"
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
			args: args{service: server.NewMetricsDumper(memory.New())},
			want: &HTTPHandler{server.NewMetricsDumper(memory.New())},
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
		{
			"/update/gauge/1",
			"/update/gauge/1 fails: metric, parsed from URL is invalid: " +
				"not found: metric name must be a string\n",
			404,
		},
		{
			"/update/WRONG/someMetric/1",
			"/update/WRONG/someMetric/1 fails: parsed metric from URL is invalid: " +
				"incorrect request: only counter and gauge types are allowed\n",
			400,
		},
		{
			"/update/counter/someMetric/1.0",
			"/update/counter/someMetric/1.0 fails: parsed metric from URL is invalid: " +
				"incorrect request: metric has invalid value\n",
			400,
		},
		{
			"/update/counter/someMetric",
			"/update/counter/someMetric fails: metric value is empty\n",
			404,
		},
		{
			"/update/counter/someMetric/9223372036854775808",
			"/update/counter/someMetric/9223372036854775808 fails: " +
				"parsed metric from URL is invalid: incorrect request: metric has invalid value\n",
			400,
		},
		{
			"/update/counter/someMetric/string",
			"/update/counter/someMetric/string fails: unable to set value for metric: " +
				"incorrect request: invalid value <string>\n",
			400,
		},
	}
	rep := memory.New()
	serv := server.NewMetricsDumper(rep)
	handler := NewHTTPHandler(serv)
	ts := httptest.NewServer(http.HandlerFunc(handler.DumpMetric))
	defer ts.Close()

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
		{
			"/value/wrong/pi",
			"/value/wrong/pi fails: metric, parsed from URL is invalid: " +
				"incorrect request: only counter and gauge types are allowed\n",
			400,
		},
		{"/value/gauge/wrong", "/value/gauge/wrong fails: not found: \n", 404},
		{"/value/counter/wrong", "/value/counter/wrong fails: not found: \n", 404},
		{"/value/counter", "/value/counter fails: not found: incorrect URL\n", 404},
		{"/value/gauge", "/value/gauge fails: not found: incorrect URL\n", 404},
	}
	m1, _ := model.NewMetric().FromValues("mainQuestion", model.MetricTypeCounter, int64(42))
	m2, _ := model.NewMetric().FromValues("pi", model.MetricTypeGauge, 3.14)
	repository := memory.New()
	_ = repository.Add(context.TODO(), m1)
	_ = repository.Add(context.TODO(), m2)

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
			method: http.MethodPost, url: "/update", contentType: constants.ContentTypeJSON,
			body:         `{"id":"mainQuestion", "type":"counter", "delta":42}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"mainQuestion", "type":"counter", "delta":42}`,
		},
		{
			method: http.MethodPost, url: "/update", contentType: constants.ContentTypeJSON,
			body:         `{"id":"mainQuestion", "type":"counter", "delta":42}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"mainQuestion", "type":"counter", "delta":84}`,
		},
		{
			method: http.MethodPost, url: "/update", contentType: constants.ContentTypeJSON,
			body:         `{"type":"counter", "delta":42}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/update", contentType: constants.ContentTypeJSON,
			body:         `{"id":"42","type":"counter", "delta":42}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/update", contentType: constants.ContentTypeJSON,
			body:         `{"id":"m42","type":"wrong", "delta":42}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/update", contentType: constants.ContentTypeJSON,
			body:         `{"id":"m42","type":"counter", "delta":42.5}`,
			expectedCode: http.StatusInternalServerError,
		},
		{
			method: http.MethodPost, url: "/update", contentType: constants.ContentTypeJSON,
			body:         `{"id":"m42","type":"counter"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/update", contentType: constants.ContentTypeJSON,
			body:         `{"id":"m42"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/update", contentType: constants.ContentTypeJSON,
			body:         `{"type":"counter"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/update", contentType: constants.ContentTypeJSON,
			body:         `{"delta":42.5}`,
			expectedCode: http.StatusInternalServerError,
		},
		{
			method: http.MethodPost, url: "/update", contentType: constants.ContentTypeJSON,
			body:         ``,
			expectedCode: http.StatusInternalServerError,
		},
		{
			method: http.MethodPost, url: "/update", contentType: constants.ContentTypeJSON,
			body:         `{"id":"m42","type":"counter", "delta":42, "value":3.14}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/update", contentType: constants.ContentTypeJSON,
			body:         `{"id":"m42","type":"wrong", "delta":"42"}`,
			expectedCode: http.StatusInternalServerError,
		},
		{
			method: http.MethodPost, url: "/update", contentType: constants.ContentTypeJSON,
			body:         `{"id":"pi", "type":"gauge", "value":3.14}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"pi", "type":"gauge", "value":3.14}`,
		},
		{
			method: http.MethodPost, url: "/update", contentType: constants.ContentTypeJSON,
			body:         `{"id":"pi", "type":"gauge", "value":3.1415926}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"pi", "type":"gauge", "value":3.1415926}`,
		},
		{
			method: http.MethodPost, url: "/update", contentType: constants.ContentTypeJSON,
			body:         `{"id":"pi", "type":"gauge", "delta":3}`,
			expectedCode: http.StatusBadRequest,
		},
	}

	repository := memory.New()
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
			method: http.MethodPost, url: "/value", contentType: constants.ContentTypeJSON,
			body:         `{"id":"m42", "type":"counter"}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"m42", "type":"counter", "delta":42}`,
		},
		{
			method: http.MethodPost, url: "/value", contentType: constants.ContentTypeJSON,
			body:         `{"type":"counter"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/value", contentType: constants.ContentTypeJSON,
			body:         `{"id":"42","type":"counter"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/value", contentType: constants.ContentTypeJSON,
			body:         `{"id":"m42","type":"wrong"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/value", contentType: constants.ContentTypeJSON,
			body:         `{"id":"m42"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			method: http.MethodPost, url: "/value", contentType: constants.ContentTypeJSON,
			body:         `{"id":"wrong","type":"counter"}`,
			expectedCode: http.StatusNotFound,
		},
		{
			method: http.MethodPost, url: "/value", contentType: constants.ContentTypeJSON,
			body:         `{"id":"pi", "type":"gauge"}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"pi", "type":"gauge", "value":3.14}`,
		},
	}

	repository := memory.New()
	m1, _ := model.NewMetric().FromValues("m42", model.MetricTypeCounter, int64(42))
	m2, _ := model.NewMetric().FromValues("pi", model.MetricTypeGauge, 3.14)
	_ = repository.Add(context.TODO(), m1)
	_ = repository.Add(context.TODO(), m2)

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
	t.Helper()

	var request *http.Request
	var err error
	if body != nil {
		request, err = http.NewRequest(method, ts.URL+path, bytes.NewBufferString(*body))
	} else {
		request, err = http.NewRequest(method, ts.URL+path, http.NoBody)
	}
	require.NoError(t, err)
	request.Header.Set(constants.KeyContentType, contentType)

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
			method: http.MethodGet, url: "/",
			expectedStatus: http.StatusOK,
			body:           "<html>\n\t<body>\n\t\t<p>m42(counter): 42</p>\n\t</body>\n</html>",
			expectedBody:   "<html>\n\t<body>\n\t\t<p>m42(counter): 42</p>\n\t</body>\n</html>",
		},
	}

	repository := memory.New()
	m1, _ := model.NewMetric().FromValues("m42", model.MetricTypeCounter, int64(42))
	_ = repository.Add(context.TODO(), m1)

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

func TestFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		want    model.Metric
	}{
		{
			name:    "valid short JSON for empty gauge",
			json:    `{"id":"m1", "type":"gauge"}}`,
			wantErr: false,
			want:    model.Metric{Name: "m1", Type: model.MetricTypeGauge},
		},
		{
			name:    "valid short JSON for empty counter",
			json:    `{"id":"m1", "type":"counter"}`,
			wantErr: false,
			want:    model.Metric{Name: "m1", Type: model.MetricTypeCounter},
		},
		{
			name:    "valid JSON for gauge -- float",
			json:    `{"id":"m1", "type":"gauge", "value": 10.0}`,
			want:    model.Metric{Name: "m1", Type: model.MetricTypeGauge, Value: func(f float64) *float64 { return &f }(10)},
			wantErr: false,
		},
		{
			name:    "valid JSON for gauge -- int",
			json:    `{"id":"m1", "type":"gauge", "value": 10}`,
			want:    model.Metric{Name: "m1", Type: model.MetricTypeGauge, Value: func(f float64) *float64 { return &f }(10)},
			wantErr: false,
		},
		{
			name:    "valid JSON for counter -- int",
			json:    `{"id":"m1", "type":"counter", "delta": 10}`,
			wantErr: false,
			want:    model.Metric{Name: "m1", Type: model.MetricTypeCounter, Delta: func(f int64) *int64 { return &f }(10)},
		},
		{
			name:    "invalid JSON for counter -- int string",
			json:    `{"id":"m1", "type":"counter", "delta": "10"}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON for counter -- float string",
			json:    `{"id":"m1", "type":"counter", "delta": "10.0"}`,
			wantErr: true,
		},
		{
			name:    "valid JSON for gauge -- float string",
			json:    `{"id":"m1", "type":"gauge", "value": "10.0"}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON for counter -- delta and value",
			json:    `{"id":"m1", "type":"counter", "delta": 10, "value": 10.0}`,
			wantErr: true,
		},
		{
			name:    "valid JSON for gauge -- delta and value",
			json:    `{"id":"m1", "type":"gauge", "value": 10.0, "delta": 10}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON -- no name, no value #1",
			json:    `{"type":"counter"}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON -- no name, no value #2",
			json:    `{"type":"gauge"}`,
			wantErr: true,
		},
		{
			name:    "invalid short JSON -- invalid type",
			json:    `{"id":"m1", "type":"invalid", "delta": 10}`,
			wantErr: true,
		},
		{
			name:    "invalid short JSON -- invalid type",
			json:    `{"id":"m1", "type":"invalid", "value": 10}`,
			wantErr: true,
		},
		{
			name:    "invalid short JSON -- invalid name for gauge",
			json:    `{"id":"1", "type":"gauge", "value": 10}`,
			wantErr: true,
		},
		{
			name:    "invalid short JSON -- invalid name for counter",
			json:    `{"id":"1", "type":"counter", "delta": 10}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON for gauge -- string value",
			json:    `{"id":"m1", "type":"gauge", "value": "value"}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON for counter -- string value",
			json:    `{"id":"m1", "type":"counter", "delta": "value"}`,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := bytes.NewBufferString(test.json)
			m, err := fromJSON(buf)
			if !test.wantErr {
				require.NoError(t, err)
				assert.Equal(t, test.want.Name, m.Name)
				assert.Equal(t, test.want.Type, m.Type)
				assert.Equal(t, test.want.ActualValue(), m.ActualValue())
				return
			}
			assert.Error(t, err)
		})
	}
}
