package service

import (
	"github.com/alant1t/metricscoll/internal/customerror"
	"github.com/alant1t/metricscoll/internal/repo"
	"strconv"
	"strings"
)

type MetricsDumper struct {
	repo repo.Repository
}

func NewMetricsDumper(repo repo.Repository) *MetricsDumper {
	return &MetricsDumper{repo: repo}
}

func (d *MetricsDumper) DumpMetric(rawMetric string) error {
	metric, err := parseURL(rawMetric)
	if err != nil {
		return err
	}
	if metric.Value == nil {
		return &customerror.InvalidArgumentError{RawMetric: rawMetric}
	}
	d.repo.Store(metric)
	return nil
}

func (d *MetricsDumper) GetMetric(rawMetric string) (repo.Metric, error) {
	m, err := parseURL(rawMetric)
	if err != nil {
		return repo.Metric{}, err
	}

	res, err := d.repo.Get(m)
	if err != nil {
		return repo.Metric{}, err
	}
	return res, nil
}

func parseURL(rawMetric string) (repo.Metric, error) {
	parts := strings.Split(rawMetric, "/")
	if len(parts) != 5 {
		return repo.Metric{},
			&customerror.NotFoundError{RawMetric: rawMetric}
	}

	mType := &parts[2]
	mName := &parts[3]
	mValue := &parts[4]
	if *mType == "" || *mName == "" || *mValue == "" {
		return repo.Metric{},
			&customerror.InvalidArgumentError{RawMetric: rawMetric}
	}

	if *mType == "gauge" {
		fValue, err := strconv.ParseFloat(*mValue, 64)
		if err != nil {
			return repo.Metric{},
				&customerror.InvalidArgumentError{RawMetric: rawMetric}
		}
		return repo.Metric{Type: repo.MetricTypeGauge, Name: *mName, Value: fValue}, nil
	} else if *mType == "counter" {
		iValue, err := strconv.Atoi(*mValue)
		if err != nil {
			return repo.Metric{},
				&customerror.InvalidArgumentError{RawMetric: rawMetric}
		}
		return repo.Metric{Type: repo.MetricTypeCounter, Name: *mName, Value: iValue}, nil
	}
	return repo.Metric{},
		&customerror.InvalidArgumentError{RawMetric: rawMetric}
}
