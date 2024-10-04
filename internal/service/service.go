package service

import "github.com/alant1t/metricscoll/internal/repo"

type Service interface {
	DumpMetric(string) error
}

type MetricsDumper struct {
	repo repo.Repository
}

func NewMetricsDumper(repo repo.Repository) *MetricsDumper {
	// TODO: почему тут нужно возвращать адрес?
	return &MetricsDumper{repo: repo}
}

func (d *MetricsDumper) DumpMetric(rawMetric string) error {
	var metric string
	var err error
	if metric, err = parseURL(rawMetric); err != nil {
		return err
	}

	if err = d.repo.Store(metric); err != nil {
		return err
	}

	return nil
}

func parseURL(metricURL string) (string, error) {
	return metricURL, nil
}
