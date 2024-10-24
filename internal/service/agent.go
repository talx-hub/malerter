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

const MetricCount = 29

const (
    MetricAlloc         repo.MetricName = "Alloc"
    MetricBuckHashSys   repo.MetricName = "BuckHashSys"
    MetricFrees         repo.MetricName = "Frees"
    MetricGCCPUFraction repo.MetricName = "GCCPUFraction"
    MetricGCSys         repo.MetricName = "GCSys"
    MetricHeapAlloc     repo.MetricName = "HeapAlloc"
    MetricHeapIdle      repo.MetricName = "HeapIdle"
    MetricHeapInuse     repo.MetricName = "HeapInuse"
    MetricHeapObjects   repo.MetricName = "HeapObjects"
    MetricHeapReleased  repo.MetricName = "HeapReleased"
    MetricHeapSys       repo.MetricName = "HeapSys"
    MetricLastGC        repo.MetricName = "LastGC"
    MetricLookups       repo.MetricName = "Lookups"
    MetricMCacheInuse   repo.MetricName = "MCacheInuse"
    MetricMCacheSys     repo.MetricName = "MCacheSys"
    MetricMSpanInuse    repo.MetricName = "MSpanInuse"
    MetricMSpanSys      repo.MetricName = "MSpanSys"
    MetricMallocs       repo.MetricName = "Mallocs"
    MetricNextGC        repo.MetricName = "NextGC"
    MetricNumForcedGC   repo.MetricName = "NumForcedGC"
    MetricNumGC         repo.MetricName = "NumGC"
    MetricOtherSys      repo.MetricName = "OtherSys"
    MetricPauseTotalNs  repo.MetricName = "PauseTotalNs"
    MetricStackInuse    repo.MetricName = "StackInuse"
    MetricStackSys      repo.MetricName = "StackSys"
    MetricSys           repo.MetricName = "Sys"
    MetricTotalAlloc    repo.MetricName = "TotalAlloc"
    MetricRandomValue   repo.MetricName = "RandomValue"
    MetricPollCount     repo.MetricName = "PollCount"
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
    var metrics = make([]repo.Metric, MetricCount)

    metrics[0] = repo.NewMetric(MetricAlloc, repo.MetricTypeGauge, float64(memStats.Alloc))
    metrics[1] = repo.NewMetric(MetricBuckHashSys, repo.MetricTypeGauge, float64(memStats.BuckHashSys))
    metrics[2] = repo.NewMetric(MetricFrees, repo.MetricTypeGauge, float64(memStats.Frees))
    metrics[3] = repo.NewMetric(MetricGCCPUFraction, repo.MetricTypeGauge, memStats.GCCPUFraction)
    metrics[4] = repo.NewMetric(MetricGCSys, repo.MetricTypeGauge, float64(memStats.GCSys))
    metrics[5] = repo.NewMetric(MetricHeapAlloc, repo.MetricTypeGauge, float64(memStats.HeapAlloc))
    metrics[6] = repo.NewMetric(MetricHeapIdle, repo.MetricTypeGauge, float64(memStats.HeapIdle))
    metrics[7] = repo.NewMetric(MetricHeapInuse, repo.MetricTypeGauge, float64(memStats.HeapInuse))
    metrics[8] = repo.NewMetric(MetricHeapObjects, repo.MetricTypeGauge, float64(memStats.HeapObjects))
    metrics[9] = repo.NewMetric(MetricHeapReleased, repo.MetricTypeGauge, float64(memStats.HeapReleased))
    metrics[10] = repo.NewMetric(MetricHeapSys, repo.MetricTypeGauge, float64(memStats.HeapSys))
    metrics[11] = repo.NewMetric(MetricLastGC, repo.MetricTypeGauge, float64(memStats.LastGC))
    metrics[12] = repo.NewMetric(MetricLookups, repo.MetricTypeGauge, float64(memStats.Lookups))
    metrics[13] = repo.NewMetric(MetricMCacheInuse, repo.MetricTypeGauge, float64(memStats.MCacheInuse))
    metrics[14] = repo.NewMetric(MetricMCacheSys, repo.MetricTypeGauge, float64(memStats.MCacheSys))
    metrics[15] = repo.NewMetric(MetricMSpanInuse, repo.MetricTypeGauge, float64(memStats.MSpanInuse))
    metrics[16] = repo.NewMetric(MetricMSpanSys, repo.MetricTypeGauge, float64(memStats.MSpanSys))
    metrics[17] = repo.NewMetric(MetricMallocs, repo.MetricTypeGauge, float64(memStats.Mallocs))
    metrics[18] = repo.NewMetric(MetricNextGC, repo.MetricTypeGauge, float64(memStats.NextGC))
    metrics[19] = repo.NewMetric(MetricNumForcedGC, repo.MetricTypeGauge, float64(memStats.NumForcedGC))
    metrics[20] = repo.NewMetric(MetricNumGC, repo.MetricTypeGauge, float64(memStats.NumGC))
    metrics[21] = repo.NewMetric(MetricOtherSys, repo.MetricTypeGauge, float64(memStats.OtherSys))
    metrics[22] = repo.NewMetric(MetricPauseTotalNs, repo.MetricTypeGauge, float64(memStats.PauseTotalNs))
    metrics[23] = repo.NewMetric(MetricStackInuse, repo.MetricTypeGauge, float64(memStats.StackInuse))
    metrics[24] = repo.NewMetric(MetricStackSys, repo.MetricTypeGauge, float64(memStats.StackSys))
    metrics[25] = repo.NewMetric(MetricSys, repo.MetricTypeGauge, float64(memStats.Sys))
    metrics[26] = repo.NewMetric(MetricTotalAlloc, repo.MetricTypeGauge, float64(memStats.TotalAlloc))
    metrics[27] = repo.NewMetric(MetricRandomValue, repo.MetricTypeGauge, randomValue)
    metrics[28] = repo.NewMetric(MetricPollCount, repo.MetricTypeCounter, int64(1))

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
