package handlers

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/memory"
)

func testRequest(t *testing.T,
	handler http.HandlerFunc,
	method,
	path,
	contentType string,
	body *string, metric ...string) (*http.Response, string) {
	t.Helper()

	var name, mType, val string
	switch len(metric) {
	case 3:
		val = metric[2]
		fallthrough
	case 2:
		mType = metric[0]
		name = metric[1]
	}

	path, err := url.JoinPath(path, metric...)
	require.NoError(t, err)

	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, path, bytes.NewBufferString(*body))
	} else {
		r = httptest.NewRequest(method, path, http.NoBody)
	}
	r.Header.Set(constants.KeyContentType, contentType)
	chiCtx := chi.NewRouteContext()
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))
	chiCtx.URLParams.Add("name", name)
	chiCtx.URLParams.Add("type", mType)
	chiCtx.URLParams.Add("val", val)

	w := httptest.NewRecorder()
	handler(w, r)
	response := w.Result()
	defer func() {
		err := response.Body.Close()
		require.NoError(t, err)
	}()

	respBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	return response, string(respBody)
}

func TestNewHTTPHandler(t *testing.T) {
	type args struct {
		storage Storage
	}

	lg, _ := logger.New(constants.LogLevelDefault)

	tests := []struct {
		name string
		args args
		want *HTTPHandler
	}{
		{
			name: "simple constructor test #0",
			args: args{nil},
			want: &HTTPHandler{nil, lg},
		},
		{
			name: "simple constructor test #1",
			args: args{storage: nil},
			want: &HTTPHandler{nil, lg},
		},
		{
			name: "simple constructor test #2",
			args: args{storage: memory.New(lg, nil)},
			want: &HTTPHandler{memory.New(lg, nil), lg},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHTTPHandler(tt.args.storage, lg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHTTPHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPHandler_DumpMetric(t *testing.T) {
	type test struct {
		mType  string
		mName  string
		mVal   string
		want   string
		status int
	}

	tests := []test{
		{"counter", "someMetric", "123", "", 200},
		{"gauge", "someMetric", "123.1", "", 200},
		{"gauge", "someMetric", "1", "", 200},
		{
			"gauge", "", "1",
			"/update/gauge/1 fails: metric, constructed from values is incorrect: " +
				"not found: metric name must be not empty\n",
			404,
		},
		{
			"WRONG", "someMetric", "1",
			"/update/WRONG/someMetric/1 fails: metric, constructed from values is incorrect: " +
				"incorrect request: only counter and gauge types are allowed\n",
			400,
		},
		{
			"counter", "someMetric", "1.0",
			"/update/counter/someMetric/1.0 fails: metric, constructed from values is incorrect: " +
				"incorrect request: metric has invalid value\n",
			400,
		},
		{
			"counter", "someMetric", "9223372036854775808",
			"/update/counter/someMetric/9223372036854775808 fails: metric, " +
				"constructed from values is incorrect: incorrect request: metric has invalid value\n",
			400,
		},
		{
			"counter", "someMetric", "string",
			"/update/counter/someMetric/string fails: unable to set value:" +
				" incorrect request: invalid value <string>\n",
			400,
		},
	}
	lg, _ := logger.New(constants.LogLevelDefault)
	rep := memory.New(lg, nil)
	handler := NewHTTPHandler(rep, lg).DumpMetric

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			resp, got := testRequest(
				t, handler, http.MethodPost, "/update",
				"", nil, tt.mType, tt.mName, tt.mVal)
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
		mType  string
		mName  string
		want   string
		status int
	}

	tests := []test{
		{"counter", "mainQuestion", "42", 200},
		{"gauge", "pi", "3.14", 200},
		{
			"wrong", "pi",
			"/value/wrong/pi fails: metric, constructed from values is incorrect: " +
				"incorrect request: only counter and gauge types are allowed\n",
			400,
		},
		{"gauge", "wrong",
			"/value/gauge/wrong fails: DB op failed: on attempt #0 error occurred: not found: \n",
			404},
		{"counter", "wrong",
			"/value/counter/wrong fails: DB op failed: on attempt #0 error occurred: not found: \n",
			404},
		{"counter", "", "/value/counter fails: metric, constructed from values is incorrect: " +
			"not found: metric name must be not empty\n", 404},
		{"gauge", "", "/value/gauge fails: metric, constructed from values is incorrect: " +
			"not found: metric name must be not empty\n", 404},
	}
	m1, _ := model.NewMetric().FromValues("mainQuestion", model.MetricTypeCounter, int64(42))
	m2, _ := model.NewMetric().FromValues("pi", model.MetricTypeGauge, 3.14)
	lg, _ := logger.New(constants.LogLevelDefault)
	repository := memory.New(lg, nil)
	_ = repository.Add(context.TODO(), m1)
	_ = repository.Add(context.TODO(), m2)

	handler := NewHTTPHandler(repository, lg).GetMetric

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			resp, got := testRequest(
				t, handler, http.MethodGet,
				"/value", "", nil,
				tt.mType, tt.mName)
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

	lg, _ := logger.New(constants.LogLevelDefault)
	repository := memory.New(lg, nil)
	handler := NewHTTPHandler(repository, lg)

	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			resp, got := testRequest(t, handler.DumpMetricJSON, test.method, test.url, test.contentType, &test.body)
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

func TestHTTPHandler_DumpMetricList(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		expectedCode int
	}{
		{
			name: "batch",
			body: `[ 
{"id":"pi", "type":"gauge", "value":3}, 
{"id":"m42","type":"counter", "delta":42}]`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "empty batch",
			body:         ``,
			expectedCode: http.StatusInternalServerError,
		},
	}

	lg, _ := logger.New(constants.LogLevelDefault)
	repository := memory.New(lg, nil)
	handler := NewHTTPHandler(repository, lg)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, _ := testRequest(t,
				handler.DumpMetricList,
				http.MethodPost, "/updates",
				constants.ContentTypeJSON,
				&test.body)
			assert.Equal(t, test.expectedCode, resp.StatusCode)
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

	lg, _ := logger.New(constants.LogLevelDefault)
	repository := memory.New(lg, nil)
	m1, _ := model.NewMetric().FromValues("m42", model.MetricTypeCounter, int64(42))
	m2, _ := model.NewMetric().FromValues("pi", model.MetricTypeGauge, 3.14)
	_ = repository.Add(context.TODO(), m1)
	_ = repository.Add(context.TODO(), m2)

	handler := NewHTTPHandler(repository, lg)

	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			resp, got := testRequest(t, handler.GetMetricJSON, test.method, test.url, test.contentType, &test.body)
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

	lg, _ := logger.New(constants.LogLevelDefault)
	repository := memory.New(lg, nil)
	m1, _ := model.NewMetric().FromValues("m42", model.MetricTypeCounter, int64(42))
	_ = repository.Add(context.TODO(), m1)

	handler := NewHTTPHandler(repository, lg)

	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			resp, got := testRequest(t, handler.GetAll, test.method, test.url, "", &test.body)
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

func TestExtractJSON(t *testing.T) {
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
			m, err := extractJSON(buf)
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
