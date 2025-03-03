package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValid(t *testing.T) {
	tests := []struct {
		name    string
		typeStr string
		want    bool
	}{
		{
			name:    "gauge metric type",
			typeStr: "gauge",
			want:    true,
		},
		{
			name:    "counter metric type",
			typeStr: "counter",
			want:    true,
		},
		{
			name:    "wrong metric type",
			typeStr: "wrong",
			want:    false,
		},
		{
			name:    "empty metric type",
			typeStr: "",
			want:    false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if ok := MetricType(test.typeStr).IsValid(); ok != test.want {
				t.Errorf("%s expected be %v, got %v", test.typeStr, test.want, ok)
			}
		})
	}
}

func TestMetricSetValue(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    Metric
		wantErr bool
	}{
		{
			name:  "int64 value type",
			value: int64(10),
			want: Metric{
				Delta: func(i int64) *int64 { return &i }(10),
				Value: nil,
			},
			wantErr: false,
		},
		{
			name:  "float64 value type",
			value: float64(10),
			want: Metric{
				Delta: nil,
				Value: func(f float64) *float64 { return &f }(10),
			},
			wantErr: false,
		},
		{
			name:  "string to int64 value type",
			value: "10",
			want: Metric{
				Delta: func(i int64) *int64 { return &i }(10),
				Value: func(f float64) *float64 { return &f }(10),
			},
			wantErr: false,
		},
		{
			name:  "string to float64 value type",
			value: "10.0",
			want: Metric{
				Delta: nil,
				Value: func(f float64) *float64 { return &f }(10),
			},
			wantErr: false,
		},
		{
			name:    "random string",
			value:   "random",
			want:    Metric{Delta: nil, Value: nil},
			wantErr: true,
		},
		{
			name:    "int value type",
			value:   10,
			want:    Metric{Delta: nil, Value: nil},
			wantErr: true,
		},
		{
			name:    "pointer to int64 value type",
			value:   func(i int64) *int64 { return &i }(10),
			want:    Metric{Delta: nil, Value: nil},
			wantErr: true,
		},
		{
			name:    "pointer to float64 value type",
			value:   func(f float64) *float64 { return &f }(10.0),
			want:    Metric{Delta: nil, Value: nil},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := NewMetric()
			err := m.setValue(test.value)

			if !test.wantErr {
				require.NoError(t, err)
				assert.Equal(t, test.want.Delta, m.Delta)
				assert.Equal(t, test.want.Value, m.Value)
				return
			}
			assert.Error(t, err)
		})
	}
}

func TestMetricClean(t *testing.T) {
	tests := []struct {
		name  string
		dirty Metric
		want  Metric
	}{
		{
			name: "gauge clean",
			dirty: Metric{
				Type:  MetricTypeGauge,
				Delta: nil,
				Value: func(f float64) *float64 { return &f }(10),
			},
			want: Metric{
				Type:  MetricTypeGauge,
				Delta: nil,
				Value: func(f float64) *float64 { return &f }(10),
			},
		},
		{
			name: "counter clean",
			dirty: Metric{
				Type:  MetricTypeCounter,
				Delta: func(i int64) *int64 { return &i }(10),
			},
			want: Metric{
				Type:  MetricTypeCounter,
				Delta: func(i int64) *int64 { return &i }(10),
			},
		},
		{
			name: "gauge with extra field",
			dirty: Metric{
				Type:  MetricTypeGauge,
				Delta: func(i int64) *int64 { return &i }(10),
				Value: func(f float64) *float64 { return &f }(10),
			},
			want: Metric{
				Type:  MetricTypeGauge,
				Delta: nil,
				Value: func(f float64) *float64 { return &f }(10),
			},
		},
		{
			name: "counter with extra field",
			dirty: Metric{
				Type:  MetricTypeCounter,
				Delta: func(i int64) *int64 { return &i }(10),
				Value: func(f float64) *float64 { return &f }(10),
			},
			want: Metric{
				Type:  MetricTypeCounter,
				Delta: func(i int64) *int64 { return &i }(10),
				Value: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.dirty.clean()
			assert.Equal(t, test.want, test.dirty)
		})
	}
}

func TestMetricIsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		metric Metric
		want   bool
	}{
		{
			name:   "empty",
			metric: Metric{},
			want:   true,
		},
		{
			name: "has delta",
			metric: Metric{
				Delta: func(i int64) *int64 { return &i }(10),
			},
			want: false,
		},
		{
			name: "has value",
			metric: Metric{
				Value: func(f float64) *float64 { return &f }(10),
			},
			want: false,
		},
		{
			name: "has both",
			metric: Metric{
				Delta: func(i int64) *int64 { return &i }(10),
				Value: func(f float64) *float64 { return &f }(10),
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, test.metric.IsEmpty())
		})
	}
}

func TestMetricIsValid(t *testing.T) {
	tests := []struct {
		name    string
		metric  Metric
		wantErr bool
	}{
		{
			name: "valid gauge",
			metric: Metric{
				Type:  MetricTypeGauge,
				Name:  "name",
				Delta: nil,
				Value: func(f float64) *float64 { return &f }(10),
			},
			wantErr: false,
		},
		{
			name: "valid counter",
			metric: Metric{
				Type:  MetricTypeCounter,
				Name:  "name",
				Delta: func(i int64) *int64 { return &i }(10),
				Value: nil,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			metric: Metric{
				Type:  MetricTypeGauge,
				Name:  "",
				Delta: nil,
				Value: func(f float64) *float64 { return &f }(10),
			},
			wantErr: true,
		},
		{
			name: "invalid name",
			metric: Metric{
				Type:  MetricTypeGauge,
				Name:  "42",
				Delta: nil,
				Value: func(f float64) *float64 { return &f }(10),
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			metric: Metric{
				Type:  MetricType("invalid"),
				Name:  "name",
				Delta: nil,
				Value: func(f float64) *float64 { return &f }(10),
			},
			wantErr: true,
		},
		{
			name: "empty gauge",
			metric: Metric{
				Type:  MetricTypeGauge,
				Name:  "name",
				Delta: nil,
				Value: nil,
			},
			wantErr: false,
		},
		{
			name: "empty counter",
			metric: Metric{
				Type:  MetricTypeCounter,
				Name:  "name",
				Delta: nil,
				Value: nil,
			},
			wantErr: false,
		},
		{
			name: "invalid value for gauge #1",
			metric: Metric{
				Type: MetricTypeGauge,
				Name: "name", Delta: func(i int64) *int64 { return &i }(10),
				Value: func(f float64) *float64 { return &f }(10),
			},
			wantErr: true,
		},
		{
			name: "invalid value for gauge #2",
			metric: Metric{
				Type:  MetricTypeGauge,
				Name:  "name",
				Delta: func(i int64) *int64 { return &i }(10), Value: nil,
			},
			wantErr: true,
		},
		{
			name: "invalid value for counter #1",
			metric: Metric{
				Type:  MetricTypeCounter,
				Name:  "name",
				Delta: func(i int64) *int64 { return &i }(10),
				Value: func(f float64) *float64 { return &f }(10),
			},
			wantErr: true,
		},
		{
			name: "invalid value for counter #2",
			metric: Metric{
				Type:  MetricTypeCounter,
				Name:  "name",
				Delta: nil,
				Value: func(f float64) *float64 { return &f }(10),
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.metric.CheckValid()
			if !test.wantErr {
				require.NoError(t, err)
				return
			}
			assert.Error(t, err)
		})
	}
}

func TestGetActualValue(t *testing.T) {
	tests := []struct {
		name   string
		metric Metric
		want   any
	}{
		{
			name:   "gauge",
			metric: Metric{Type: MetricTypeGauge, Value: func(f float64) *float64 { return &f }(10)},
			want:   10.0,
		},
		{
			name:   "counter",
			metric: Metric{Type: MetricTypeCounter, Delta: func(i int64) *int64 { return &i }(10)},
			want:   int64(10),
		},
		{
			name:   "empty gauge",
			metric: Metric{Type: MetricTypeGauge},
			want:   nil,
		},
		{
			name:   "empty counter",
			metric: Metric{Type: MetricTypeCounter},
			want:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, test.metric.ActualValue())
		})
	}
}

func TestMetricUpdate(t *testing.T) {
	tests := []struct {
		name    string
		lhs     Metric
		rhs     Metric
		wantErr bool
		want    Metric
	}{
		{
			name:    "valid gauges",
			lhs:     Metric{Name: "lhs", Type: MetricTypeGauge, Value: func(f float64) *float64 { return &f }(5)},
			rhs:     Metric{Name: "rhs", Type: MetricTypeGauge, Value: func(f float64) *float64 { return &f }(10)},
			want:    Metric{Type: MetricTypeGauge, Value: func(f float64) *float64 { return &f }(10)},
			wantErr: false,
		},
		{
			name:    "valid counters",
			lhs:     Metric{Name: "lhs", Type: MetricTypeCounter, Delta: func(i int64) *int64 { return &i }(10)},
			rhs:     Metric{Name: "rhs", Type: MetricTypeCounter, Delta: func(i int64) *int64 { return &i }(11)},
			want:    Metric{Type: MetricTypeCounter, Delta: func(i int64) *int64 { return &i }(21)},
			wantErr: false,
		},
		// TODO: handle OF
		/*
			{
				name:    "counters overflow",
				lhs:     Metric{Name: "lhs", Type: MetricTypeCounter, Delta: func(i int64) *int64 { return &i }(math.MaxInt64)},
				rhs:     Metric{Name: "rhs", Type: MetricTypeCounter, Delta: func(i int64) *int64 { return &i }(math.MaxInt64)},
				wantErr: true,
			},
		*/
		{
			name:    "empty gauge + valid gauge",
			lhs:     Metric{Name: "lhs", Type: MetricTypeGauge, Value: nil},
			rhs:     Metric{Name: "rhs", Type: MetricTypeGauge, Value: func(f float64) *float64 { return &f }(10)},
			wantErr: true,
		},
		{
			name:    "valid gauge + empty gauge",
			lhs:     Metric{Name: "lhs", Type: MetricTypeGauge, Value: func(f float64) *float64 { return &f }(10)},
			rhs:     Metric{Name: "rhs", Type: MetricTypeGauge, Value: nil},
			wantErr: true,
		},
		{
			name:    "empty gauge + empty gauge",
			lhs:     Metric{Name: "lhs", Type: MetricTypeGauge, Value: nil},
			rhs:     Metric{Name: "rhs", Type: MetricTypeGauge, Value: nil},
			wantErr: true,
		},
		{
			name:    "valid counter + empty counter",
			lhs:     Metric{Name: "lhs", Type: MetricTypeCounter, Delta: func(i int64) *int64 { return &i }(10)},
			rhs:     Metric{Name: "rhs", Type: MetricTypeCounter, Delta: nil},
			wantErr: true,
		},
		{
			name:    "empty counter + valid counter",
			lhs:     Metric{Name: "lhs", Type: MetricTypeCounter, Delta: nil},
			rhs:     Metric{Name: "rhs", Type: MetricTypeCounter, Delta: func(i int64) *int64 { return &i }(10)},
			wantErr: true,
		},
		{
			name:    "empty counter + empty counter",
			lhs:     Metric{Name: "lhs", Type: MetricTypeCounter, Delta: nil},
			rhs:     Metric{Name: "rhs", Type: MetricTypeCounter, Delta: nil},
			wantErr: true,
		},
		{
			name:    "gauge + counter",
			lhs:     Metric{Name: "lhs", Type: MetricTypeGauge, Value: func(f float64) *float64 { return &f }(10)},
			rhs:     Metric{Name: "rhs", Type: MetricTypeCounter, Delta: func(i int64) *int64 { return &i }(10)},
			wantErr: true,
		},
		{
			name:    "counter + gauge",
			lhs:     Metric{Name: "lhs", Type: MetricTypeCounter, Delta: func(i int64) *int64 { return &i }(10)},
			rhs:     Metric{Name: "rhs", Type: MetricTypeGauge, Value: func(f float64) *float64 { return &f }(10)},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.lhs.Update(test.rhs)
			if !test.wantErr {
				require.NoError(t, err)
				assert.Equal(t, test.want.ActualValue(), test.lhs.ActualValue())
				return
			}
			assert.Error(t, err)
		})
	}
}

func TestFromValues(t *testing.T) {
	type values struct {
		name  string
		mType MetricType
		value any
	}
	tests := []struct {
		name    string
		values  values
		want    Metric
		wantErr bool
	}{
		{
			name:    "create valid gauge, value float",
			values:  values{"m1", MetricTypeGauge, 10.0},
			want:    Metric{Name: "m1", Type: MetricTypeGauge, Value: func(f float64) *float64 { return &f }(10)},
			wantErr: false,
		},
		{
			name:    "create valid gauge, value -- string",
			values:  values{"m1", MetricTypeGauge, "10.0"},
			want:    Metric{Name: "m1", Type: MetricTypeGauge, Value: func(f float64) *float64 { return &f }(10)},
			wantErr: false,
		},
		{
			name:    "create valid counter, value -- int64",
			values:  values{"m1", MetricTypeCounter, int64(10)},
			want:    Metric{Name: "m1", Type: MetricTypeCounter, Delta: func(f int64) *int64 { return &f }(10)},
			wantErr: false,
		},
		{
			name:    "create valid counter, value -- string",
			values:  values{"m1", MetricTypeCounter, "10"},
			want:    Metric{Name: "m1", Type: MetricTypeCounter, Delta: func(f int64) *int64 { return &f }(10)},
			wantErr: false,
		},
		{
			name:    "invalid name for gauge #1",
			values:  values{"1", MetricTypeGauge, "10.0"},
			wantErr: true,
		},
		{
			name:    "invalid name for counter #1",
			values:  values{"1", MetricTypeCounter, int64(10)},
			wantErr: true,
		},
		{
			name:    "invalid name for gauge #2 -- empty",
			values:  values{"", MetricTypeGauge, "10.0"},
			wantErr: true,
		},
		{
			name:    "invalid name for counter #2 -- empty",
			values:  values{"", MetricTypeCounter, int64(10)},
			wantErr: true,
		},
		{
			name:    "invalid type #1",
			values:  values{"m1", MetricType("invalid"), "10.0"},
			wantErr: true,
		},
		{
			name:    "invalid type #2 -- empty type",
			values:  values{"m1", MetricType(""), "10.0"},
			wantErr: true,
		},
		{
			name:    "invalid counter, value -- int",
			values:  values{"m1", MetricTypeCounter, 10},
			wantErr: true,
		},
		{
			name:    "invalid counter, value -- float",
			values:  values{"m1", MetricTypeCounter, 10.0},
			wantErr: true,
		},
		{
			name:    "invalid counter, value -- nil",
			values:  values{"m1", MetricTypeCounter, nil},
			wantErr: true,
		},
		{
			name:    "invalid gauge, value -- int64",
			values:  values{"m1", MetricTypeGauge, int64(10)},
			wantErr: true,
		},
		{
			name:    "invalid gauge, value -- int",
			values:  values{"m1", MetricTypeGauge, 10},
			wantErr: true,
		},
		{
			name:    "invalid gauge, value -- nil",
			values:  values{"m1", MetricTypeGauge, nil},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m, err := NewMetric().FromValues(test.values.name, test.values.mType, test.values.value)
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
