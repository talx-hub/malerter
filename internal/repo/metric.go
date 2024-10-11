package repo

import (
	"fmt"
)

const MetricCount = 29

type MetricName string

const (
	MetricAlloc         MetricName = "Alloc"
	MetricBuckHashSys   MetricName = "BuckHashSys"
	MetricFrees         MetricName = "Frees"
	MetricGCCPUFraction MetricName = "GCCPUFraction"
	MetricGCSys         MetricName = "GCSys"
	MetricHeapAlloc     MetricName = "HeapAlloc"
	MetricHeapIdle      MetricName = "HeapIdle"
	MetricHeapInuse     MetricName = "HeapInuse"
	MetricHeapObjects   MetricName = "HeapObjects"
	MetricHeapReleased  MetricName = "HeapReleased"
	MetricHeapSys       MetricName = "HeapSys"
	MetricLastGC        MetricName = "LastGC"
	MetricLookups       MetricName = "Lookups"
	MetricMCacheInuse   MetricName = "MCacheInuse"
	MetricMCacheSys     MetricName = "MCacheSys"
	MetricMSpanInuse    MetricName = "MSpanInuse"
	MetricMSpanSys      MetricName = "MSpanSys"
	MetricMallocs       MetricName = "Mallocs"
	MetricNextGC        MetricName = "NextGC"
	MetricNumForcedGC   MetricName = "NumForcedGC"
	MetricNumGC         MetricName = "NumGC"
	MetricOtherSys      MetricName = "OtherSys"
	MetricPauseTotalNs  MetricName = "PauseTotalNs"
	MetricStackInuse    MetricName = "StackInuse"
	MetricStackSys      MetricName = "StackSys"
	MetricSys           MetricName = "Sys"
	MetricTotalAlloc    MetricName = "TotalAlloc"
	MetricRandomValue   MetricName = "RandomValue"
	MetricPollCount     MetricName = "PollCount"
)

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

func NewMetric(name MetricName, value any) Metric {
	var mType MetricType
	if name == MetricPollCount {
		mType = MetricTypeCounter
	} else {
		mType = MetricTypeGauge
	}
	return Metric{Type: mType, Name: name.String(), Value: value}
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
