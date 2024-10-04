package service

import (
	"github.com/alant1t/metricscoll/internal/customerror"
	"github.com/alant1t/metricscoll/internal/repo"
	"strconv"
	"strings"
)

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
	var metric Metric
	var err error
	if metric, err = parseURL(rawMetric); err != nil {
		return err
	}

	if err = d.repo.Store(metric); err != nil {
		return err
	}

	return nil
}

func parseURL(rawMetric string) (Metric, error) {
	parts := strings.Split(rawMetric, "/")
	if len(parts) != 4 {
		return Metric{},
			&customerror.IvalidArgumentError{RawMetric: rawMetric}
	}

	if parts[1] == "gauge" {
		fValue, err := strconv.ParseFloat(parts[3], 64)
		if err != nil {
			return Metric{},
				&customerror.IvalidArgumentError{RawMetric: rawMetric}
		}
		return Metric{mType: parts[1], name: parts[2], fValue: fValue}, nil
	} else if parts[1] == "counter" {
		iValue, err := strconv.Atoi(parts[3])
		if err != nil {
			return Metric{},
				&customerror.IvalidArgumentError{RawMetric: rawMetric}
		}
		return Metric{mType: parts[1], name: parts[2], iValue: iValue}, nil
	}
	return Metric{},
		&customerror.IvalidArgumentError{RawMetric: rawMetric}
}
