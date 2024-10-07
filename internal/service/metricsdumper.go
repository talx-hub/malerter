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
	if len(parts) < 4 {
		return repo.Metric{},
			&customerror.NotFoundError{RawMetric: rawMetric}
	}

	// только два типа метрик позволены
	mType := repo.MetricType(parts[2])
	if !mType.IsValid() {
		return repo.Metric{},
			&customerror.InvalidArgumentError{RawMetric: rawMetric}
	}

	// имя не должно быть числом
	mName := &parts[3]
	_, errF := strconv.ParseFloat(*mName, 64)
	_, errI := strconv.Atoi(*mName)
	if errF == nil || errI == nil {
		return repo.Metric{},
			&customerror.NotFoundError{RawMetric: rawMetric}
	}
	if len(parts) == 4 {
		return repo.Metric{Type: mType, Name: *mName, Value: nil}, nil
	}

	// значение должно быть числом и соответствовать типу
	mValue := &parts[4]
	iVal, iErr := strconv.Atoi(*mValue)
	if mType == repo.MetricTypeCounter && iErr == nil {
		return repo.Metric{Type: mType, Name: *mName, Value: iVal}, nil
	}

	fVal, fErr := strconv.ParseFloat(*mValue, 64)
	if mType == repo.MetricTypeGauge && fErr != nil {
		return repo.Metric{Type: mType, Name: *mName, Value: fVal}, nil
	}

	return repo.Metric{},
		&customerror.InvalidArgumentError{RawMetric: rawMetric}
}
