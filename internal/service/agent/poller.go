package agent

import (
	"log"
	"math/rand/v2"
	"runtime"

	"github.com/talx-hub/malerter/internal/model"
)

const MetricCount = 29

const (
	MetricAlloc         string = "Alloc"
	MetricBuckHashSys   string = "BuckHashSys"
	MetricFrees         string = "Frees"
	MetricGCCPUFraction string = "GCCPUFraction"
	MetricGCSys         string = "GCSys"
	MetricHeapAlloc     string = "HeapAlloc"
	MetricHeapIdle      string = "HeapIdle"
	MetricHeapInuse     string = "HeapInuse"
	MetricHeapObjects   string = "HeapObjects"
	MetricHeapReleased  string = "HeapReleased"
	MetricHeapSys       string = "HeapSys"
	MetricLastGC        string = "LastGC"
	MetricLookups       string = "Lookups"
	MetricMCacheInuse   string = "MCacheInuse"
	MetricMCacheSys     string = "MCacheSys"
	MetricMSpanInuse    string = "MSpanInuse"
	MetricMSpanSys      string = "MSpanSys"
	MetricMallocs       string = "Mallocs"
	MetricNextGC        string = "NextGC"
	MetricNumForcedGC   string = "NumForcedGC"
	MetricNumGC         string = "NumGC"
	MetricOtherSys      string = "OtherSys"
	MetricPauseTotalNs  string = "PauseTotalNs"
	MetricStackInuse    string = "StackInuse"
	MetricStackSys      string = "StackSys"
	MetricSys           string = "Sys"
	MetricTotalAlloc    string = "TotalAlloc"
	MetricRandomValue   string = "RandomValue"
	MetricPollCount     string = "PollCount"
)

type Poller struct {
	storage Storage
}

func (p *Poller) update() {
	metrics := collect()
	p.store(metrics)
}

func collect() []model.Metric {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	var randomValue = rand.Float64()
	var m = make([]model.Metric, MetricCount)

	m[0], _ = model.NewMetric().FromValues(MetricAlloc, model.MetricTypeGauge, float64(memStats.Alloc))
	m[1], _ = model.NewMetric().FromValues(MetricBuckHashSys, model.MetricTypeGauge, float64(memStats.BuckHashSys))
	m[2], _ = model.NewMetric().FromValues(MetricFrees, model.MetricTypeGauge, float64(memStats.Frees))
	m[3], _ = model.NewMetric().FromValues(MetricGCCPUFraction, model.MetricTypeGauge, memStats.GCCPUFraction)
	m[4], _ = model.NewMetric().FromValues(MetricGCSys, model.MetricTypeGauge, float64(memStats.GCSys))
	m[5], _ = model.NewMetric().FromValues(MetricHeapAlloc, model.MetricTypeGauge, float64(memStats.HeapAlloc))
	m[6], _ = model.NewMetric().FromValues(MetricHeapIdle, model.MetricTypeGauge, float64(memStats.HeapIdle))
	m[7], _ = model.NewMetric().FromValues(MetricHeapInuse, model.MetricTypeGauge, float64(memStats.HeapInuse))
	m[8], _ = model.NewMetric().FromValues(MetricHeapObjects, model.MetricTypeGauge, float64(memStats.HeapObjects))
	m[9], _ = model.NewMetric().FromValues(MetricHeapReleased, model.MetricTypeGauge, float64(memStats.HeapReleased))
	m[10], _ = model.NewMetric().FromValues(MetricHeapSys, model.MetricTypeGauge, float64(memStats.HeapSys))
	m[11], _ = model.NewMetric().FromValues(MetricLastGC, model.MetricTypeGauge, float64(memStats.LastGC))
	m[12], _ = model.NewMetric().FromValues(MetricLookups, model.MetricTypeGauge, float64(memStats.Lookups))
	m[13], _ = model.NewMetric().FromValues(MetricMCacheInuse, model.MetricTypeGauge, float64(memStats.MCacheInuse))
	m[14], _ = model.NewMetric().FromValues(MetricMCacheSys, model.MetricTypeGauge, float64(memStats.MCacheSys))
	m[15], _ = model.NewMetric().FromValues(MetricMSpanInuse, model.MetricTypeGauge, float64(memStats.MSpanInuse))
	m[16], _ = model.NewMetric().FromValues(MetricMSpanSys, model.MetricTypeGauge, float64(memStats.MSpanSys))
	m[17], _ = model.NewMetric().FromValues(MetricMallocs, model.MetricTypeGauge, float64(memStats.Mallocs))
	m[18], _ = model.NewMetric().FromValues(MetricNextGC, model.MetricTypeGauge, float64(memStats.NextGC))
	m[19], _ = model.NewMetric().FromValues(MetricNumForcedGC, model.MetricTypeGauge, float64(memStats.NumForcedGC))
	m[20], _ = model.NewMetric().FromValues(MetricNumGC, model.MetricTypeGauge, float64(memStats.NumGC))
	m[21], _ = model.NewMetric().FromValues(MetricOtherSys, model.MetricTypeGauge, float64(memStats.OtherSys))
	m[22], _ = model.NewMetric().FromValues(MetricPauseTotalNs, model.MetricTypeGauge, float64(memStats.PauseTotalNs))
	m[23], _ = model.NewMetric().FromValues(MetricStackInuse, model.MetricTypeGauge, float64(memStats.StackInuse))
	m[24], _ = model.NewMetric().FromValues(MetricStackSys, model.MetricTypeGauge, float64(memStats.StackSys))
	m[25], _ = model.NewMetric().FromValues(MetricSys, model.MetricTypeGauge, float64(memStats.Sys))
	m[26], _ = model.NewMetric().FromValues(MetricTotalAlloc, model.MetricTypeGauge, float64(memStats.TotalAlloc))
	m[27], _ = model.NewMetric().FromValues(MetricRandomValue, model.MetricTypeGauge, randomValue)
	m[28], _ = model.NewMetric().FromValues(MetricPollCount, model.MetricTypeCounter, int64(1))

	return m
}

func (p *Poller) store(metrics []model.Metric) {
	for _, m := range metrics {
		if p.storage.Add(m) != nil {
			log.Println("error during storing of metric")
		}
	}
}
