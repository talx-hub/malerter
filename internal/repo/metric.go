package repo

import (
	"strconv"
	"strings"
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

type Metric struct {
	Type  string
	Name  string
	Value any
}

func NewMetric(name MetricName, value any) Metric {
	var mType string
	if name == MetricPollCount {
		mType = "counter"
	} else {
		mType = "gauge"
	}
	return Metric{Type: mType, Name: name.String(), Value: value}
}

func (m *Metric) String() string {
	var value string
	if m.Type == "gauge" {
		value = strconv.FormatFloat(m.Value.(float64), 'f', 2, 64)
	} else {
		value = strconv.Itoa(m.Value.(int))
	}

	var mSlice = []string{m.Type, m.Name, value}
	return strings.Join(mSlice, "/")
}

func (m *Metric) Update(other Metric) {
	if m.Type == "gauge" {
		m.Value = other.Value
	} else {
		updated := m.Value.(int) + other.Value.(int)
		m.Value = updated
	}
}
