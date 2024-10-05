package service

import (
	r "github.com/alant1t/metricscoll/internal/repo"
	"math/rand"
	"net/http"
	"runtime"
)

const host = "http://localhost:8080"

type Agent struct {
	repo r.Repository
}

func NewAgent(repo r.Repository) *Agent {
	return &Agent{repo: repo}
}

func (a *Agent) Update() error {
	metrics := collect()
	a.store(metrics)
	return nil
}

func (a *Agent) Send() error {
	metrics := a.get()
	urls := convertToURLs(metrics)
	send(urls)
	return nil
}

func collect() []r.Metric {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	var randomValue = rand.Float64()
	var metrics = make([]r.Metric, r.MetricCount, r.MetricCount)

	metrics[0] = r.NewMetric(r.MetricAlloc, float64(memStats.Alloc))
	metrics[1] = r.NewMetric(r.MetricBuckHashSys, float64(memStats.BuckHashSys))
	metrics[2] = r.NewMetric(r.MetricFrees, float64(memStats.Frees))
	metrics[3] = r.NewMetric(r.MetricGCCPUFraction, memStats.GCCPUFraction)
	metrics[4] = r.NewMetric(r.MetricGCSys, float64(memStats.GCSys))
	metrics[5] = r.NewMetric(r.MetricHeapAlloc, float64(memStats.HeapAlloc))
	metrics[6] = r.NewMetric(r.MetricHeapIdle, float64(memStats.HeapIdle))
	metrics[7] = r.NewMetric(r.MetricHeapInuse, float64(memStats.HeapInuse))
	metrics[8] = r.NewMetric(r.MetricHeapObjects, float64(memStats.HeapObjects))
	metrics[9] = r.NewMetric(r.MetricHeapReleased, float64(memStats.HeapReleased))
	metrics[10] = r.NewMetric(r.MetricHeapSys, float64(memStats.HeapSys))
	metrics[11] = r.NewMetric(r.MetricLastGC, float64(memStats.LastGC))
	metrics[12] = r.NewMetric(r.MetricLookups, float64(memStats.Lookups))
	metrics[13] = r.NewMetric(r.MetricMCacheInuse, float64(memStats.MCacheInuse))
	metrics[14] = r.NewMetric(r.MetricMCacheSys, float64(memStats.MCacheSys))
	metrics[15] = r.NewMetric(r.MetricMSpanInuse, float64(memStats.MSpanInuse))
	metrics[16] = r.NewMetric(r.MetricMSpanSys, float64(memStats.MSpanSys))
	metrics[17] = r.NewMetric(r.MetricMallocs, float64(memStats.Mallocs))
	metrics[18] = r.NewMetric(r.MetricNextGC, float64(memStats.NextGC))
	metrics[19] = r.NewMetric(r.MetricNumForcedGC, float64(memStats.NumForcedGC))
	metrics[20] = r.NewMetric(r.MetricNumGC, float64(memStats.NumGC))
	metrics[21] = r.NewMetric(r.MetricOtherSys, float64(memStats.OtherSys))
	metrics[22] = r.NewMetric(r.MetricPauseTotalNs, float64(memStats.PauseTotalNs))
	metrics[23] = r.NewMetric(r.MetricStackInuse, float64(memStats.StackInuse))
	metrics[24] = r.NewMetric(r.MetricStackSys, float64(memStats.StackSys))
	metrics[25] = r.NewMetric(r.MetricSys, float64(memStats.Sys))
	metrics[26] = r.NewMetric(r.MetricTotalAlloc, float64(memStats.TotalAlloc))
	metrics[27] = r.NewMetric(r.MetricRandomValue, randomValue)
	metrics[28] = r.NewMetric(r.MetricPollCount, 1)

	return metrics
}

func (a *Agent) store(metrics []r.Metric) {
	for _, m := range metrics {
		a.repo.Store(m)
	}
}

func (a *Agent) get() []r.Metric {
	return a.repo.GetAll()
}

func convertToURLs(metrics []r.Metric) []string {
	var urls []string
	for _, m := range metrics {
		url := host + "/update/" + m.String()
		urls = append(urls, url)
	}
	return urls
}

func send(urls []string) {
	for _, url := range urls {
		http.Post(url, "text/plain", nil)
	}
}
