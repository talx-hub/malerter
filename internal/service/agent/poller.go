package agent

import (
	"context"
	"fmt"
	"math/rand/v2"
	"runtime"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
)

const runtimeMetricCount = 29

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

const (
	MetricTotalMemory string = "TotalMemory"
	MetricFreeMemory  string = "FreeMemory"
)

type Poller struct {
	storage Storage
	log     *logger.ZeroLogger
}

func (p *Poller) update() {
	runtimeMetrics := collectRuntime()
	p.store(runtimeMetrics)

	psutilMetrics, err := collectPsutil()
	if err != nil {
		p.log.Error().Err(err).Msg("failed to collect psutil metrics")
		return
	}
	p.store(psutilMetrics)
}

func collectRuntime() []model.Metric {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	var randomValue = rand.Float64()
	var m = make([]model.Metric, runtimeMetricCount)

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

func collectPsutil() ([]model.Metric, error) {
	memory, err := mem.VirtualMemory()
	if err != nil {
		return nil,
			fmt.Errorf("failed to collect memory utilization: %w", err)
	}
	totalMem := float64(memory.Total)
	freeMem := float64(memory.Free)

	totalMemMetric, _ := model.NewMetric().FromValues(MetricTotalMemory, model.MetricTypeGauge, totalMem)
	freeMemMetric, _ := model.NewMetric().FromValues(MetricFreeMemory, model.MetricTypeGauge, freeMem)

	const compareWithLastCall time.Duration = 0
	const percpu = true
	cpuUtil, err := cpu.Percent(compareWithLastCall, percpu)
	if err != nil {
		return nil,
			fmt.Errorf("failed to collect CPU utilization: %w", err)
	}

	cpuUtilMetrics := make([]model.Metric, len(cpuUtil))
	for i, v := range cpuUtil {
		nameMetric := "CPUutilization" + strconv.Itoa(i+1)
		m, _ := model.NewMetric().FromValues(nameMetric, model.MetricTypeGauge, v)
		cpuUtilMetrics[i] = m
	}

	metrics := make([]model.Metric, 0)
	metrics = append(metrics, totalMemMetric, freeMemMetric)
	metrics = append(metrics, cpuUtilMetrics...)

	return metrics, nil
}

func (p *Poller) store(metrics []model.Metric) {
	for _, m := range metrics {
		if p.storage.Add(context.TODO(), m) != nil {
			p.log.Error().Msg("error during storing of metric")
		}
	}
}
