package repo

import (
	"fmt"
)

type MetricName string

func (m MetricName) String() string {
	return string(m)
}

type MetricType string

const (
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeCounter MetricType = "counter"
)

func (t MetricType) IsValid() bool {
	return t == MetricTypeGauge || t == MetricTypeCounter
}

func (t MetricType) String() string {
	return string(t)
}

type Metric struct {
	Type  MetricType
	Name  string
	Value any
}

func NewMetric(name MetricName, mType MetricType, value any) Metric {
	return Metric{
		Type:  mType,
		Name:  name.String(),
		Value: value,
	}
}

func (m *Metric) String() string {
	return fmt.Sprintf("%s(%s): %v",
		m.Name, m.Type.String(), m.Value)
}

func (m *Metric) ToURL() string {
	return fmt.Sprintf("%s/%s/%v", m.Type.String(), m.Name, m.Value)
}

func (m *Metric) Update(other Metric) {
	if m.Type == MetricTypeGauge {
		m.Value = other.Value
	} else {
		updated := m.Value.(int64) + other.Value.(int64)
		m.Value = updated
	}
}
