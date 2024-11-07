package repo

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/talx-hub/malerter/internal/customerror"
)

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

func NewMetric() *Metric {
	return &Metric{}
}

func (m *Metric) setValue(val any) error {
	if iVal, ok := val.(int64); ok {
		m.Delta = &iVal
		return nil
	}
	if fVal, ok := val.(float64); ok {
		m.Value = &fVal
		return nil
	}

	sVal, ok := val.(string)
	if !ok {
		return &customerror.InvalidArgumentError{
			MetricURL: m.ToURL(),
			Info:      fmt.Sprintf("invalid value <%v>", val),
		}
	}
	fVal, fErr := strconv.ParseFloat(sVal, 64)
	iVal, iErr := strconv.ParseInt(sVal, 10, 64)
	if fErr == nil {
		m.Value = &fVal
	}
	if iErr == nil {
		m.Delta = &iVal
	}
	if fErr == nil || iErr == nil {
		return nil
	}
	return &customerror.InvalidArgumentError{
		MetricURL: m.ToURL(),
		Info:      fmt.Sprintf("invalid value <%v>", val),
	}
}

// if Metric.setValue() receives string that could be converted to int,
// the Metric.setValue() method will set both m.Value and m.Delta,
// so we need to clean extra field
func (m *Metric) clean() {
	if m.Type == MetricTypeGauge && m.Delta != nil && m.Value != nil {
		m.Delta = nil
	}
	if m.Type == MetricTypeCounter && m.Value != nil && m.Delta != nil {
		m.Value = nil
	}
}

func (m *Metric) IsEmpty() bool {
	if m.Delta == nil && m.Value == nil {
		return true
	}
	return false
}

func (m *Metric) checkValid() error {
	// имя не должно быть числом
	_, errF := strconv.ParseFloat(m.Name, 64)
	_, errI := strconv.Atoi(m.Name)
	if errF == nil || errI == nil {
		return &customerror.NotFoundError{
			MetricURL: m.ToURL(),
			Info:      "metric name must be a string",
		}
	}

	// только два типа метрик позволены
	if !m.Type.IsValid() {
		return &customerror.InvalidArgumentError{
			MetricURL: m.ToURL(),
			Info:      "only counter and gauge types are allowed",
		}
	}

	// значение должно соответствовать типу
	wrongCounter := m.Type == MetricTypeCounter && m.Delta == nil
	wrongGauge := m.Type == MetricTypeGauge && m.Delta != nil && m.Value == nil
	if !m.IsEmpty() && (wrongCounter || wrongGauge) {
		return &customerror.InvalidArgumentError{
			MetricURL: m.ToURL(),
			Info:      "metric has invalid value",
		}
	}

	return nil
}

func (m *Metric) String() string {
	if fVal, ok := m.ActualValue().(float64); ok {
		fValStr := strconv.FormatFloat(fVal, 'f', 2, 64)
		return fmt.Sprintf("%s(%s): %v", m.Name, m.Type.String(), fValStr)
	}
	return fmt.Sprintf("%s(%s): %v", m.Name, m.Type.String(), m.ActualValue())
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
	}
	if m.Type == MetricTypeCounter && m.Delta != nil {
		return *m.Delta
	}
	return nil
}

func (m *Metric) FromValues(name string, t MetricType, value any) (*Metric, error) {
	m.Name = name
	m.Type = t

	if err := m.setValue(value); err != nil {
		return nil, err
	}
	m.clean()
	if err := m.checkValid(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Metric) FromURL(url string) (*Metric, error) {
	parts := strings.Split(url, "/")
	if len(parts) < 4 {
		return nil, &customerror.NotFoundError{
			MetricURL: url,
			Info:      "incorrect URL",
		}
	}

	m.Name = parts[3]
	m.Type = MetricType(parts[2])
	if len(parts) == 4 {
		if err := m.checkValid(); err != nil {
			return nil, err
		}
		return m, nil
	}

	if err := m.setValue(parts[4]); err != nil {
		return nil, err
	}
	m.clean()
	if err := m.checkValid(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Metric) FromJSON(body io.ReadCloser) (*Metric, error) {
	if err := json.NewDecoder(body).Decode(m); err != nil {
		return nil, err
	}
	m.clean()
	if err := m.checkValid(); err != nil {
		return nil, err
	}

	return m, nil
}
