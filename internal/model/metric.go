package model

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

func (m *Metric) CheckValid() error {
	if m.Name == "" {
		return &customerror.NotFoundError{
			MetricURL: m.ToURL(),
			Info:      "metric name must be not empty",
		}
	}

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
	wrongCounter := m.Type == MetricTypeCounter && (m.Value != nil || m.Delta == nil)
	wrongGauge := m.Type == MetricTypeGauge && (m.Delta != nil || m.Value == nil)
	if !m.IsEmpty() && (wrongCounter || wrongGauge) {
		return &customerror.InvalidArgumentError{
			MetricURL: m.ToURL(),
			Info:      "metric has invalid value",
		}
	}

	return nil
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

func (m *Metric) String() string {
	if m.IsEmpty() {
		return fmt.Sprintf("%s(%s): <nil>", m.Name, m.Type.String())
	}
	if m.Type == MetricTypeGauge {
		return fmt.Sprintf("%s(%s): %.2f", m.Name, m.Type.String(), m.ActualValue())
	}
	return fmt.Sprintf("%s(%s): %v", m.Name, m.Type.String(), m.ActualValue())
}

func (m *Metric) ToURL() string {
	if m.IsEmpty() {
		return fmt.Sprintf("%s/%s/<nil>", m.Type.String(), m.Name)
	}
	if m.Type == MetricTypeGauge {
		return fmt.Sprintf("%s/%s/%.2f", m.Type.String(), m.Name, m.ActualValue())
	}
	return fmt.Sprintf("%s/%s/%v", m.Type.String(), m.Name, m.ActualValue())
}

func (m *Metric) Update(other Metric) error {
	// TODO: может убрать эти проверки??? невалидные метрики вообще не должны иметь возможность быть созданными клиентским кодом
	if err := m.CheckValid(); err != nil {
		return fmt.Errorf("cannot update invalid metric: %v", err)
	}
	if err := other.CheckValid(); err != nil {
		return fmt.Errorf("rhs metric is invalid, cannot update: %v", err)
	}
	if m.IsEmpty() {
		return fmt.Errorf("lhs metric is empty, cannot update")
	}
	if other.IsEmpty() {
		return fmt.Errorf("rhs metric is empty, cannot update")
	}
	if m.Type != other.Type {
		return fmt.Errorf("lhs and rhs metrics type are different, cannot update")
	}

	if m.Type == MetricTypeGauge {
		m.Value = other.Value
	} else {
		// TODO: handle overflow
		updated := *m.Delta + *other.Delta
		m.Delta = &updated
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
	if err := m.CheckValid(); err != nil {
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
		if err := m.CheckValid(); err != nil {
			return nil, err
		}
		return m, nil
	}

	if err := m.setValue(parts[4]); err != nil {
		return nil, err
	}
	m.clean()
	if err := m.CheckValid(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Metric) FromJSON(body io.Reader) (*Metric, error) {
	if err := json.NewDecoder(body).Decode(m); err != nil {
		return nil, err
	}
	if err := m.CheckValid(); err != nil {
		return nil, err
	}

	return m, nil
}
