package agent

import (
	"log"
	"math/rand/v2"
	"runtime"

	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repo"
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
	repo repo.Repository
}

func (p *Poller) update() error {
	metrics := collect()
	p.store(metrics)
	return nil
}

func collect() []*model.Metric {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	var randomValue = rand.Float64()
	var metrics = make([]*model.Metric, MetricCount)

	metrics[0], _ = model.NewMetric().FromValues(MetricAlloc, model.MetricTypeGauge, float64(memStats.Alloc))
	metrics[1], _ = model.NewMetric().FromValues(MetricBuckHashSys, model.MetricTypeGauge, float64(memStats.BuckHashSys))
	metrics[2], _ = model.NewMetric().FromValues(MetricFrees, model.MetricTypeGauge, float64(memStats.Frees))
	metrics[3], _ = model.NewMetric().FromValues(MetricGCCPUFraction, model.MetricTypeGauge, memStats.GCCPUFraction)
	metrics[4], _ = model.NewMetric().FromValues(MetricGCSys, model.MetricTypeGauge, float64(memStats.GCSys))
	metrics[5], _ = model.NewMetric().FromValues(MetricHeapAlloc, model.MetricTypeGauge, float64(memStats.HeapAlloc))
	metrics[6], _ = model.NewMetric().FromValues(MetricHeapIdle, model.MetricTypeGauge, float64(memStats.HeapIdle))
	metrics[7], _ = model.NewMetric().FromValues(MetricHeapInuse, model.MetricTypeGauge, float64(memStats.HeapInuse))
	metrics[8], _ = model.NewMetric().FromValues(MetricHeapObjects, model.MetricTypeGauge, float64(memStats.HeapObjects))
	metrics[9], _ = model.NewMetric().FromValues(MetricHeapReleased, model.MetricTypeGauge, float64(memStats.HeapReleased))
	metrics[10], _ = model.NewMetric().FromValues(MetricHeapSys, model.MetricTypeGauge, float64(memStats.HeapSys))
	metrics[11], _ = model.NewMetric().FromValues(MetricLastGC, model.MetricTypeGauge, float64(memStats.LastGC))
	metrics[12], _ = model.NewMetric().FromValues(MetricLookups, model.MetricTypeGauge, float64(memStats.Lookups))
	metrics[13], _ = model.NewMetric().FromValues(MetricMCacheInuse, model.MetricTypeGauge, float64(memStats.MCacheInuse))
	metrics[14], _ = model.NewMetric().FromValues(MetricMCacheSys, model.MetricTypeGauge, float64(memStats.MCacheSys))
	metrics[15], _ = model.NewMetric().FromValues(MetricMSpanInuse, model.MetricTypeGauge, float64(memStats.MSpanInuse))
	metrics[16], _ = model.NewMetric().FromValues(MetricMSpanSys, model.MetricTypeGauge, float64(memStats.MSpanSys))
	metrics[17], _ = model.NewMetric().FromValues(MetricMallocs, model.MetricTypeGauge, float64(memStats.Mallocs))
	metrics[18], _ = model.NewMetric().FromValues(MetricNextGC, model.MetricTypeGauge, float64(memStats.NextGC))
	metrics[19], _ = model.NewMetric().FromValues(MetricNumForcedGC, model.MetricTypeGauge, float64(memStats.NumForcedGC))
	metrics[20], _ = model.NewMetric().FromValues(MetricNumGC, model.MetricTypeGauge, float64(memStats.NumGC))
	metrics[21], _ = model.NewMetric().FromValues(MetricOtherSys, model.MetricTypeGauge, float64(memStats.OtherSys))
	metrics[22], _ = model.NewMetric().FromValues(MetricPauseTotalNs, model.MetricTypeGauge, float64(memStats.PauseTotalNs))
	metrics[23], _ = model.NewMetric().FromValues(MetricStackInuse, model.MetricTypeGauge, float64(memStats.StackInuse))
	metrics[24], _ = model.NewMetric().FromValues(MetricStackSys, model.MetricTypeGauge, float64(memStats.StackSys))
	metrics[25], _ = model.NewMetric().FromValues(MetricSys, model.MetricTypeGauge, float64(memStats.Sys))
	metrics[26], _ = model.NewMetric().FromValues(MetricTotalAlloc, model.MetricTypeGauge, float64(memStats.TotalAlloc))
	metrics[27], _ = model.NewMetric().FromValues(MetricRandomValue, model.MetricTypeGauge, randomValue)
	metrics[28], _ = model.NewMetric().FromValues(MetricPollCount, model.MetricTypeCounter, int64(1))

	return metrics
}

func (p *Poller) store(metrics []*model.Metric) {
	for _, m := range metrics {
		if m == nil {
			log.Println("can't store empty metric")
			continue
		}
		p.repo.Store(*m)
	}
}
