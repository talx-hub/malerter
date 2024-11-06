package repo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/talx-hub/malerter/internal/customerror"
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
	Type  MetricType `json:"type"`
	Name  string     `json:"id"`
	Delta *int64     `json:"delta,omitempty"`
	Value *float64   `json:"value,omitempty"`
}

func NewMetric(name MetricName, mType MetricType, value any) Metric {
	m := Metric{
		Type:  mType,
		Name:  name.String(),
		Delta: nil,
		Value: nil,
	}
	if iVal, ok := value.(int64); ok {
		m.Delta = &iVal
		return m
	}
	if fVal, ok := value.(float64); ok {
		m.Value = &fVal
		return m
	}
	return m
}

func (m *Metric) String() string {
	return fmt.Sprintf("%s(%s): %v",
		m.Name, m.Type.String(), m.Value)
}

func (m *Metric) ToURL() string {
	if fVal, ok := m.ActualValue().(float64); ok {
		fValStr := strconv.FormatFloat(fVal, 'f', 2, 64)
		return fmt.Sprintf("%s/%s/%v", m.Type.String(), m.Name, fValStr)
	}
	return fmt.Sprintf("%s/%s/%v", m.Type.String(), m.Name, m.ActualValue())
}

func (m *Metric) Update(other Metric) {
	if m.Type == MetricTypeGauge {
		m.Value = other.Value
	} else {
		updated := *m.Delta + *other.Delta
		m.Delta = &updated
	}
}

func (m *Metric) ActualValue() any {
	if m.Type == MetricTypeGauge && m.Value != nil {
		return *m.Value
	} else if m.Type == MetricTypeCounter && m.Delta != nil {
		return *m.Delta
	} else {
		return nil
	}
}

func FromURL(url string) (Metric, error) {
	parts := strings.Split(url, "/")
	if len(parts) < 4 {
		return Metric{},
			&customerror.NotFoundError{
				MetricURL: url,
				Info:      "incorrect URL",
			}
	}

	// только два типа метрик позволены
	mType := MetricType(parts[2])
	if !mType.IsValid() {
		return Metric{},
			&customerror.InvalidArgumentError{
				MetricURL: url,
				Info:      "only counter and gauge types are allowed",
			}
	}

	// имя не должно быть числом
	mName := &parts[3]
	_, errF := strconv.ParseFloat(*mName, 64)
	_, errI := strconv.Atoi(*mName)
	if errF == nil || errI == nil {
		return Metric{},
			&customerror.NotFoundError{
				MetricURL: url,
				Info:      "metric name must be a string",
			}
	}
	if len(parts) == 4 {
		return Metric{Type: mType, Name: *mName, Value: nil}, nil
	}

	// значение должно быть числом и соответствовать типу
	mValue := &parts[4]
	iVal, iErr := strconv.ParseInt(*mValue, 10, 64)
	if mType == MetricTypeCounter && iErr == nil {
		return Metric{Type: mType, Name: *mName, Delta: &iVal}, nil
	}

	fVal, fErr := strconv.ParseFloat(*mValue, 64)
	if mType == MetricTypeGauge && fErr == nil {
		return Metric{Type: mType, Name: *mName, Value: &fVal}, nil
	}

	return Metric{},
		&customerror.InvalidArgumentError{
			MetricURL: url,
			Info:      "wrong value type for metric type",
		}
}
