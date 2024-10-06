package api

import (
	"github.com/alant1t/metricscoll/internal/repo"
	"github.com/alant1t/metricscoll/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHTTPHandler_DumpMetric(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		method string
		reqURL string
		want   want
	}{
		{
			name:   "positive test #1: counter",
			method: http.MethodPost,
			reqURL: "/update/counter/someMetric/123",
			want: want{
				code:     200,
				response: "",
			},
		},
		{
			name:   "positive test #2: gauge",
			method: http.MethodPost,
			reqURL: "/update/gauge/someMetric/123.1",
			want: want{
				code:     200,
				response: "",
			},
		},
		{
			name:   "positive test #3: int gauge",
			method: http.MethodPost,
			reqURL: "/update/gauge/someMetric/1",
			want: want{
				code:     200,
				response: "",
			},
		},
		{
			name:   "negative test #1: wrong method",
			method: http.MethodGet,
			reqURL: "/update/gauge/someMetric/1",
			want: want{
				code:     400,
				response: "only POST requests are allowed\n",
			},
		},
		{
			name:   "negative test #2: no metric name",
			method: http.MethodPost,
			reqURL: "/update/gauge/1",
			want: want{
				code:     404,
				response: "metric /update/gauge/1 not found\n",
			},
		},
		{
			name:   "negative test #3: wrong metric type",
			method: http.MethodPost,
			reqURL: "/update/gage/someMetric/1",
			want: want{
				code:     400,
				response: "metric /update/gage/someMetric/1 is incorrect\n",
			},
		},
		{
			name:   "negative test #4: wrong value type",
			method: http.MethodPost,
			reqURL: "/update/counter/someMetric/1.0",
			want: want{
				code:     400,
				response: "metric /update/counter/someMetric/1.0 is incorrect\n",
			},
		},
		{
			name:   "negative test #5: no value",
			method: http.MethodPost,
			reqURL: "/update/counter/someMetric",
			want: want{
				code:     404,
				response: "metric /update/counter/someMetric not found\n",
			},
		},
		{
			name:   "negative test #6: value overflow",
			method: http.MethodPost,
			reqURL: "/update/counter/someMetric/9223372036854775808",
			want: want{
				code:     400,
				response: "metric /update/counter/someMetric/9223372036854775808 is incorrect\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.reqURL, nil)
			w := httptest.NewRecorder()

			rep := repo.NewMemRepository()
			serv := service.NewMetricsDumper(rep)
			handler := NewHTTPHandler(serv)

			handler.DumpMetric(w, request)
			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, tt.want.response, string(resBody))
		})
	}
}

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
