package service

import (
	"github.com/talx-hub/malerter/internal/customerror"
	"github.com/talx-hub/malerter/internal/repo"
)

type MetricsDumper struct {
	repo repo.Repository
}

func NewMetricsDumper(repo repo.Repository) *MetricsDumper {
	return &MetricsDumper{repo: repo}
}

func (d *MetricsDumper) Store(metric repo.Metric) error {
	if metric.Value == nil && metric.Delta == nil {
		return &customerror.NotFoundError{MetricURL: metric.ToURL()}
	}
	d.repo.Store(metric)
	return nil
}

func (d *MetricsDumper) Get(metric repo.Metric) (repo.Metric, error) {
	res, err := d.repo.Get(metric)
	if err != nil {
		return repo.Metric{}, err
	}
	return res, nil
}

func (d *MetricsDumper) GetAll() []repo.Metric {
	return d.repo.GetAll()
}
