package agent

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
)

func TestToJSONs(t *testing.T) {
	log, err := logger.New(constants.LogLevelDefault)
	require.NoError(t, err)
	sender := Sender{host: "", log: log}

	tests := []struct {
		name    string
		metrics []model.Metric
		want    []string
	}{
		{
			name:    "empty input",
			metrics: []model.Metric{},
			want:    []string{},
		},
		{
			name: "normal metrics",
			metrics: []model.Metric{
				{Type: model.MetricTypeCounter, Name: "m42", Delta: func(i int64) *int64 { return &i }(42), Value: nil},
				{Type: model.MetricTypeGauge, Name: "pi", Delta: nil, Value: func(f float64) *float64 { return &f }(3.14)},
				{Type: model.MetricTypeGauge, Name: "a", Delta: nil, Value: func(f float64) *float64 { return &f }(3.15)},
				{Type: model.MetricTypeGauge, Name: "b", Delta: nil, Value: func(f float64) *float64 { return &f }(3.16)},
				{Type: model.MetricTypeGauge, Name: "c", Delta: nil, Value: func(f float64) *float64 { return &f }(3.17)},
				{Type: model.MetricTypeGauge, Name: "d", Delta: nil, Value: func(f float64) *float64 { return &f }(3.18)},
				{Type: model.MetricTypeGauge, Name: "e", Delta: nil, Value: func(f float64) *float64 { return &f }(3.19)},
			},
			want: []string{
				`{"id":"m42","type":"counter","delta":42}`,
				`{"id":"pi","type":"gauge","value":3.14}`,
				`{"id":"a","type":"gauge","value":3.15}`,
				`{"id":"b","type":"gauge","value":3.16}`,
				`{"id":"c","type":"gauge","value":3.17}`,
				`{"id":"d","type":"gauge","value":3.18}`,
				`{"id":"e","type":"gauge","value":3.19}`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ch := make(chan model.Metric, len(test.metrics))
			for _, m := range test.metrics {
				ch <- m
			}
			close(ch)
			jsonsCh := sender.toJSONs(ch)

			i := 0
			for json := range jsonsCh {
				assert.JSONEq(t, json, test.want[i])
				i++
			}
		})
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		name  string
		jsons []string
		want  string
	}{
		{
			name:  "empty input",
			jsons: []string{},
			want:  "[]",
		},
		{
			name:  "single metric",
			jsons: []string{`{"id":"m42","type":"counter","delta":42}`},
			want:  `[{"id":"m42","type":"counter","delta":42}]`,
		},
		{
			name: "a lot of metrics",
			jsons: []string{
				`{"id":"m42","type":"counter","delta":42}`,
				`{"id":"pi","type":"gauge","value":3.14}`,
				`{"id":"a","type":"gauge","value":3.15}`},
			want: `[{"id":"m42","type":"counter","delta":42},` +
				`{"id":"pi","type":"gauge","value":3.14},` +
				`{"id":"a","type":"gauge","value":3.15}]`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ch := make(chan string, len(test.jsons))
			for _, j := range test.jsons {
				ch <- j
			}
			close(ch)
			batch := join(ch)
			assert.JSONEq(t, batch, test.want)
		})
	}
}

func newTestSender(serverURL string, secret string, compress bool) *Sender {
	return &Sender{
		client:   &http.Client{Timeout: 1 * time.Second},
		log:      logger.NewNopLogger(),
		host:     serverURL,
		secret:   secret,
		compress: compress,
	}
}

func TestSender_toJSONs(t *testing.T) {
	s := newTestSender("", "", false)
	input := make(chan model.Metric, 2)
	input <- model.Metric{Name: "test1", Type: model.MetricTypeGauge, Value: new(float64)}
	input <- model.Metric{Name: "test2", Type: model.MetricTypeGauge, Value: new(float64)}
	close(input)

	out := s.toJSONs(input)
	results := make([]string, 0)
	for json := range out {
		results = append(results, json)
	}

	assert.Len(t, results, 2)
	assert.Contains(t, results[0], `"test1"`)
	assert.Contains(t, results[1], `"test2"`)
}

func TestSender_batch_Success(t *testing.T) {
	// Start mock server
	var receivedBody []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		receivedBody, err = io.ReadAll(r.Body)
		defer func() {
			err = r.Body.Close()
			require.NoError(t, err)
		}()
		assert.NoError(t, err)

		assert.Equal(t, constants.ContentTypeJSON, r.Header.Get(constants.KeyContentType))
	}))
	defer ts.Close()

	s := newTestSender(ts.URL, "", false)
	s.batch([]byte(`[{"id":"1"}]`), "", false, false)

	assert.Contains(t, string(receivedBody), `"id":"1"`)
}

func TestSender_batch_Compress(t *testing.T) {
	// Server that checks gzip encoding
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, constants.EncodingGzip, r.Header.Get(constants.KeyContentEncoding))
		_, err := io.ReadAll(r.Body)
		defer func() {
			err = r.Body.Close()
			require.NoError(t, err)
		}()
		assert.NoError(t, err)
	}))
	defer ts.Close()

	s := newTestSender(ts.URL, "", true)
	compressed, err := s.tryCompress([]byte(`[{"id":"2"}]`))
	require.NoError(t, err)
	s.batch(compressed, "", true, false)
}

func TestSender_batch_Signature(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sig := r.Header.Get(constants.KeyHashSHA256)
		assert.NotEmpty(t, sig)
	}))
	defer ts.Close()

	s := newTestSender(ts.URL, "super-secret", false)
	data := []byte(`[{"name":"metric"}]`)
	sig := s.trySign(data)
	s.batch(data, sig, false, false)
}

func TestSender_batch_HTTPError(t *testing.T) {
	s := newTestSender("http://localhost:9999", "", false) // Unused port to simulate error
	s.batch([]byte(`[{"bad":"json"}]`), "", false, false)
	// No panic/assert — we just check it doesn't crash
}

func TestSender_send(t *testing.T) {
	var mu sync.Mutex
	var wg sync.WaitGroup

	chMetrics := make(chan model.Metric, 1)
	chMetrics <- model.Metric{Name: "test2", Type: model.MetricTypeGauge, Value: new(float64)}
	close(chMetrics)

	chJobs := make(chan chan model.Metric, 1)
	chJobs <- chMetrics
	close(chJobs)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
	}))
	defer ts.Close()

	s := newTestSender(ts.URL, "", false)
	wg.Add(1)
	go s.send(chJobs, &mu, &wg)

	wg.Wait()
}
