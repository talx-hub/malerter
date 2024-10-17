package service

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/talx-hub/malerter/internal/config"
	"github.com/talx-hub/malerter/internal/repo"
)

type Agent struct {
	repo   repo.Repository
	config *config.Agent
}

func NewAgent(repo repo.Repository, cfg *config.Agent) *Agent {
	return &Agent{repo: repo, config: cfg}
}

func (a *Agent) Run() {
	var i = 1
	var updateToSendRatio = int(a.config.ReportInterval / a.config.PollInterval)
	for {
		if err := a.Update(); err != nil {
			if _, e := os.Stderr.WriteString(err.Error()); e != nil {
				log.Fatal(e)
			}
		}

		if i%updateToSendRatio == 0 {
			if err := a.Send(); err != nil {
				if _, e := os.Stderr.WriteString(err.Error()); e != nil {
					log.Fatal(e)
				}
			}
			i = 0
		}
		i++
		time.Sleep(a.config.PollInterval)
	}
}

func (a *Agent) Update() error {
	metrics := collect()
	a.store(metrics)
	return nil
}

func (a *Agent) Send() error {
	metrics := a.get()
	urls := convertToURLs(metrics, a.config.ServerAddress)
	send(urls)
	return nil
}

func collect() []repo.Metric {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	var randomValue = rand.Float64()
	var metrics = make([]repo.Metric, repo.MetricCount)

	metrics[0] = repo.NewMetric(repo.MetricAlloc, float64(memStats.Alloc))
	metrics[1] = repo.NewMetric(repo.MetricBuckHashSys, float64(memStats.BuckHashSys))
	metrics[2] = repo.NewMetric(repo.MetricFrees, float64(memStats.Frees))
	metrics[3] = repo.NewMetric(repo.MetricGCCPUFraction, memStats.GCCPUFraction)
	metrics[4] = repo.NewMetric(repo.MetricGCSys, float64(memStats.GCSys))
	metrics[5] = repo.NewMetric(repo.MetricHeapAlloc, float64(memStats.HeapAlloc))
	metrics[6] = repo.NewMetric(repo.MetricHeapIdle, float64(memStats.HeapIdle))
	metrics[7] = repo.NewMetric(repo.MetricHeapInuse, float64(memStats.HeapInuse))
	metrics[8] = repo.NewMetric(repo.MetricHeapObjects, float64(memStats.HeapObjects))
	metrics[9] = repo.NewMetric(repo.MetricHeapReleased, float64(memStats.HeapReleased))
	metrics[10] = repo.NewMetric(repo.MetricHeapSys, float64(memStats.HeapSys))
	metrics[11] = repo.NewMetric(repo.MetricLastGC, float64(memStats.LastGC))
	metrics[12] = repo.NewMetric(repo.MetricLookups, float64(memStats.Lookups))
	metrics[13] = repo.NewMetric(repo.MetricMCacheInuse, float64(memStats.MCacheInuse))
	metrics[14] = repo.NewMetric(repo.MetricMCacheSys, float64(memStats.MCacheSys))
	metrics[15] = repo.NewMetric(repo.MetricMSpanInuse, float64(memStats.MSpanInuse))
	metrics[16] = repo.NewMetric(repo.MetricMSpanSys, float64(memStats.MSpanSys))
	metrics[17] = repo.NewMetric(repo.MetricMallocs, float64(memStats.Mallocs))
	metrics[18] = repo.NewMetric(repo.MetricNextGC, float64(memStats.NextGC))
	metrics[19] = repo.NewMetric(repo.MetricNumForcedGC, float64(memStats.NumForcedGC))
	metrics[20] = repo.NewMetric(repo.MetricNumGC, float64(memStats.NumGC))
	metrics[21] = repo.NewMetric(repo.MetricOtherSys, float64(memStats.OtherSys))
	metrics[22] = repo.NewMetric(repo.MetricPauseTotalNs, float64(memStats.PauseTotalNs))
	metrics[23] = repo.NewMetric(repo.MetricStackInuse, float64(memStats.StackInuse))
	metrics[24] = repo.NewMetric(repo.MetricStackSys, float64(memStats.StackSys))
	metrics[25] = repo.NewMetric(repo.MetricSys, float64(memStats.Sys))
	metrics[26] = repo.NewMetric(repo.MetricTotalAlloc, float64(memStats.TotalAlloc))
	metrics[27] = repo.NewMetric(repo.MetricRandomValue, randomValue)
	metrics[28] = repo.NewMetric(repo.MetricPollCount, int64(1))

	return metrics
}

func (a *Agent) store(metrics []repo.Metric) {
	for _, m := range metrics {
		a.repo.Store(m)
	}
}

func (a *Agent) get() []repo.Metric {
	return a.repo.GetAll()
}

func convertToURLs(metrics []repo.Metric, host string) []string {
	var urls []string
	for _, m := range metrics {
		url := "http://" + host + "/update/" + m.ToURL()
		urls = append(urls, url)
	}
	return urls
}

func send(urls []string) {
	for _, url := range urls {
		response, err := http.Post(url, "text/plain", nil)
		if err != nil {
			continue
		}
		if err := response.Body.Close(); err != nil {
			log.Fatal(err)
		}
	}
}
