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
	return d.repo.Store(metric)
}

func (d *MetricsDumper) Get(key string) (model.Metric, error) {
	res, err := d.repo.Get(key)
	if err != nil {
		return model.Metric{}, err
	}
	return res, nil
}

func (d *MetricsDumper) GetAll() []model.Metric {
	return d.repo.GetAll()
}
