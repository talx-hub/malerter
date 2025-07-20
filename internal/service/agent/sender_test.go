package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/service/server/logger"
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

/*
Func TestSend(t *testing.T) {
	tests := []struct {
		name    string
		metrics []model.Metric
		want    []string
	}{
		{
			name:    "empty",
			metrics: []model.Metric{},
			want:    []string{},
		},
		{
			name: "no error",
			metrics: []model.Metric{
				{
					Delta: func(i int64) *int64 { return &i }(42),
					Value: nil,
					Type:  "counter",
					Name:  "m42",
				},
				{
					Delta: nil,
					Value: func(i float64) *float64 { return &i }(3.14),
					Type:  "gauge",
					Name:  "pi",
				},
			},
			want: []string{
				`{"id":"m42","type":"counter","delta":42}`,
				`{"id":"pi","type":"gauge","value":3.14}`,
			},
		},
	}
	// FIXME: эта городуха вообще норм? :DDD
	storage := ""
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if body, err := io.ReadAll(r.Body); err != nil {
			log.Println(err)
		} else {
			storage = string(body)
		}
		if err := r.Body.Close(); err != nil {
			log.Println(err)
		}
	}))
	defer testServer.Close()

	client := testServer.Client()
	sender := Sender{
		client:   client,
		storage:  memory.New(),
		host:     testServer.URL,
		compress: false,
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, m := range test.metrics {
				err := sender.storage.Add(context.TODO(), m)
				require.NoError(t, err)
			}
			sender.send()
			for _, str := range test.want {
				assert.True(t, strings.Contains(storage, str))
			}
		})
	}
}.
*/
