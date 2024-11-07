package service

import (
	"log"
	"math/rand/v2"
	"runtime"

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

func collect() []*repo.Metric {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	var randomValue = rand.Float64()
	var metrics = make([]*repo.Metric, MetricCount)

	metrics[0], _ = repo.NewMetric().FromValues(MetricAlloc, repo.MetricTypeGauge, float64(memStats.Alloc))
	metrics[1], _ = repo.NewMetric().FromValues(MetricBuckHashSys, repo.MetricTypeGauge, float64(memStats.BuckHashSys))
	metrics[2], _ = repo.NewMetric().FromValues(MetricFrees, repo.MetricTypeGauge, float64(memStats.Frees))
	metrics[3], _ = repo.NewMetric().FromValues(MetricGCCPUFraction, repo.MetricTypeGauge, memStats.GCCPUFraction)
	metrics[4], _ = repo.NewMetric().FromValues(MetricGCSys, repo.MetricTypeGauge, float64(memStats.GCSys))
	metrics[5], _ = repo.NewMetric().FromValues(MetricHeapAlloc, repo.MetricTypeGauge, float64(memStats.HeapAlloc))
	metrics[6], _ = repo.NewMetric().FromValues(MetricHeapIdle, repo.MetricTypeGauge, float64(memStats.HeapIdle))
	metrics[7], _ = repo.NewMetric().FromValues(MetricHeapInuse, repo.MetricTypeGauge, float64(memStats.HeapInuse))
	metrics[8], _ = repo.NewMetric().FromValues(MetricHeapObjects, repo.MetricTypeGauge, float64(memStats.HeapObjects))
	metrics[9], _ = repo.NewMetric().FromValues(MetricHeapReleased, repo.MetricTypeGauge, float64(memStats.HeapReleased))
	metrics[10], _ = repo.NewMetric().FromValues(MetricHeapSys, repo.MetricTypeGauge, float64(memStats.HeapSys))
	metrics[11], _ = repo.NewMetric().FromValues(MetricLastGC, repo.MetricTypeGauge, float64(memStats.LastGC))
	metrics[12], _ = repo.NewMetric().FromValues(MetricLookups, repo.MetricTypeGauge, float64(memStats.Lookups))
	metrics[13], _ = repo.NewMetric().FromValues(MetricMCacheInuse, repo.MetricTypeGauge, float64(memStats.MCacheInuse))
	metrics[14], _ = repo.NewMetric().FromValues(MetricMCacheSys, repo.MetricTypeGauge, float64(memStats.MCacheSys))
	metrics[15], _ = repo.NewMetric().FromValues(MetricMSpanInuse, repo.MetricTypeGauge, float64(memStats.MSpanInuse))
	metrics[16], _ = repo.NewMetric().FromValues(MetricMSpanSys, repo.MetricTypeGauge, float64(memStats.MSpanSys))
	metrics[17], _ = repo.NewMetric().FromValues(MetricMallocs, repo.MetricTypeGauge, float64(memStats.Mallocs))
	metrics[18], _ = repo.NewMetric().FromValues(MetricNextGC, repo.MetricTypeGauge, float64(memStats.NextGC))
	metrics[19], _ = repo.NewMetric().FromValues(MetricNumForcedGC, repo.MetricTypeGauge, float64(memStats.NumForcedGC))
	metrics[20], _ = repo.NewMetric().FromValues(MetricNumGC, repo.MetricTypeGauge, float64(memStats.NumGC))
	metrics[21], _ = repo.NewMetric().FromValues(MetricOtherSys, repo.MetricTypeGauge, float64(memStats.OtherSys))
	metrics[22], _ = repo.NewMetric().FromValues(MetricPauseTotalNs, repo.MetricTypeGauge, float64(memStats.PauseTotalNs))
	metrics[23], _ = repo.NewMetric().FromValues(MetricStackInuse, repo.MetricTypeGauge, float64(memStats.StackInuse))
	metrics[24], _ = repo.NewMetric().FromValues(MetricStackSys, repo.MetricTypeGauge, float64(memStats.StackSys))
	metrics[25], _ = repo.NewMetric().FromValues(MetricSys, repo.MetricTypeGauge, float64(memStats.Sys))
	metrics[26], _ = repo.NewMetric().FromValues(MetricTotalAlloc, repo.MetricTypeGauge, float64(memStats.TotalAlloc))
	metrics[27], _ = repo.NewMetric().FromValues(MetricRandomValue, repo.MetricTypeGauge, randomValue)
	metrics[28], _ = repo.NewMetric().FromValues(MetricPollCount, repo.MetricTypeCounter, int64(1))

	return metrics
}

func (p *Poller) store(metrics []*repo.Metric) {
	for _, m := range metrics {
		if m == nil {
			log.Println("can't store empty metric")
			continue
		}
		p.repo.Store(*m)
	}
}
