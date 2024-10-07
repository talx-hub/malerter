package api

import (
	"github.com/alant1t/metricscoll/internal/customerror"
	"github.com/alant1t/metricscoll/internal/repo"
	"github.com/alant1t/metricscoll/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
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
			args: args{service: service.NewMetricsDumper(nil)},
			want: &HTTPHandler{service.NewMetricsDumper(nil)},
		},
		{
			name: "simple constructor test #2",
			args: args{service: service.NewMetricsDumper(repo.NewMemRepository())},
			want: &HTTPHandler{service.NewMetricsDumper(repo.NewMemRepository())},
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

type test struct {
	url    string
	want   string
	status int
}

func TestHTTPHandler_DumpMetric(t *testing.T) {
	tests := []test{
		{"/update/counter/someMetric/123", "", 200},
		{"/update/gauge/someMetric/123.1", "", 200},
		{"/update/gauge/someMetric/1", "", 200},
		{"/update/gauge/1", "metric /update/gauge/1 not found\n", 404},
		{"/update/WRONG/someMetric/1", "metric /update/WRONG/someMetric/1 is incorrect\n", 400},
		{"/update/counter/someMetric/1.0", "metric /update/counter/someMetric/1.0 is incorrect\n", 400},
		{"/update/counter/someMetric", "metric /update/counter/someMetric not found\n", 404},
		{"/update/counter/someMetric/9223372036854775808", "metric /update/counter/someMetric/9223372036854775808 is incorrect\n", 400},
		{"/update/counter/someMetric/string", "metric /update/counter/someMetric/string is incorrect\n", 400},
	}
	rep := repo.NewMemRepository()
	serv := service.NewMetricsDumper(rep)
	handler := NewHTTPHandler(serv)
	ts := httptest.NewServer(http.HandlerFunc(handler.DumpMetric))
	defer ts.Close()

	wrongMethodTest := test{
		url:    "/update/gauge/someMetric/1",
		want:   "only POST requests are allowed\n",
		status: 400,
	}
	t.Run("wrong method test", func(t *testing.T) {
		resp, got := testRequest(t, ts, http.MethodGet, wrongMethodTest.url)
		assert.Equal(t, wrongMethodTest.status, resp.StatusCode)
		assert.Equal(t, wrongMethodTest.want, got)
		if err := resp.Body.Close(); err != nil {
			log.Fatal(err)
		}
	})

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			resp, got := testRequest(t, ts, http.MethodPost, tt.url)
			assert.Equal(t, tt.status, resp.StatusCode)
			assert.Equal(t, tt.want, got)
			if err := resp.Body.Close(); err != nil {
				log.Fatal(err)
			}
		})
	}
}

type mockRepo struct {
	d map[string]repo.Metric
}

func (r *mockRepo) Store(repo.Metric) {}
func (r *mockRepo) Get(m repo.Metric) (repo.Metric, error) {
	dummyKey := m.Type.String() + m.Name
	if mm, found := r.d[dummyKey]; found {
		return mm, nil
	}
	return repo.Metric{},
		&customerror.NotFoundError{RawMetric: m.String()}
}
func (r *mockRepo) GetAll() []repo.Metric {
	var metrics []repo.Metric
	for _, m := range r.d {
		metrics = append(metrics, m)
	}
	return metrics
}

func TestHTTPHandler_GetMetric(t *testing.T) {
	tests := []test{
		{"/value/counter/mainQuestion", "42", 200},
		{"/value/gauge/pi", "3.14", 200},
		{"/value/wrong/pi", "metric /value/wrong/pi is incorrect\n", 400},
		{"/value/gauge/wrong", "metric wrong(gauge): <nil> not found\n", 404},
		{"/value/counter/wrong", "metric wrong(counter): <nil> not found\n", 404},
		{"/value/counter", "metric /value/counter not found\n", 404},
		{"/value/gauge", "metric /value/gauge not found\n", 404},
	}
	m := map[string]repo.Metric{
		"countermainQuestion": {Type: repo.MetricTypeCounter, Name: "mainQuestion", Value: int64(42)},
		"gaugepi":             {Type: repo.MetricTypeGauge, Name: "pi", Value: 3.14},
	}
	mock := mockRepo{d: m}
	serv := service.NewMetricsDumper(&mock)
	handler := NewHTTPHandler(serv)
	ts := httptest.NewServer(http.HandlerFunc(handler.GetMetric))
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			resp, got := testRequest(t, ts, http.MethodGet, tt.url)
			assert.Equal(t, tt.status, resp.StatusCode)
			assert.Equal(t, tt.want, got)
			if err := resp.Body.Close(); err != nil {
				log.Fatal(err)
			}
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server,
	method, path string) (*http.Response, string) {

	request, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(request)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(body)
}
