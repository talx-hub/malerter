package agent

import (
	"github.com/talx-hub/malerter/internal/repo"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talx-hub/malerter/internal/model"
)

func TestGet(t *testing.T) {
	storage := repo.NewMemRepository()
	m1, _ := model.NewMetric().FromValues("m42", model.MetricTypeCounter, int64(42))
	m2, _ := model.NewMetric().FromValues("pi", model.MetricTypeGauge, 3.14)
	_ = storage.Store(*m1)
	_ = storage.Store(*m2)
	sender := Sender{repo: storage, host: ""}
	got := sender.get()
	require.Len(t, got, 2)
	assert.Contains(t, got, *m1)
	assert.Contains(t, got, *m2)
}

func TestConvertToJSONs(t *testing.T) {
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
			},
			want: []string{
				`{"id":"m42","type":"counter","delta":42}`,
				`{"id":"pi","type":"gauge","value":3.14}`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := convertToJSONs(test.metrics)
			require.Equal(t, len(test.want), len(got))
			for i, json := range got {
				assert.JSONEq(t, json, test.want[i])
			}
		})
	}
}

func TestSend(t *testing.T) {
	tests := []struct {
		name  string
		jsons []string
	}{
		{
			name:  "empty",
			jsons: []string{},
		},
		{
			name: "no error",
			jsons: []string{
				`{"id":"m42","type":"counter","delta":42}`,
				`{"id":"pi","type":"gauge","value":3.14}`,
			},
		},
	}
	// FIXME: эта городуха вообще норм? :DDD
	storage := make([]string, 0)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if body, err := io.ReadAll(r.Body); err != nil {
			log.Println(err)
		} else {
			storage = append(storage, string(body))
		}
		if err := r.Body.Close(); err != nil {
			log.Println(err)
		}
	}))
	defer testServer.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			send(testServer.URL, test.jsons)
			require.Equal(t, len(test.jsons), len(storage))
			for i, json := range test.jsons {
				assert.JSONEq(t, json, storage[i])
			}
		})
	}
}
