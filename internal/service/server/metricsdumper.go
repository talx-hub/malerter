package server

import (
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repo"
)

type MetricsDumper struct {
	repo repo.Repository
}

func NewMetricsDumper(repo repo.Repository) *MetricsDumper {
	return &MetricsDumper{repo: repo}
}

func (d *MetricsDumper) Store(metric model.Metric) error {
	d.repo.Store(metric)
	return nil
}

func (d *MetricsDumper) Get(metric model.Metric) (model.Metric, error) {
	res, err := d.repo.Get(metric)
	if err != nil {
		return model.Metric{}, err
	}
	return res, nil
}

func (d *MetricsDumper) GetAll() []model.Metric {
	return d.repo.GetAll()
}
