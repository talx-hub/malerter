package server

import (
	"github.com/talx-hub/malerter/internal/model"
)

type Storage interface {
	Add(metric model.Metric) error
	Find(key string) (model.Metric, error)
	Get() []model.Metric
}

type MetricsDumper struct {
	Storage
}

func NewMetricsDumper(repo Storage) *MetricsDumper {
	return &MetricsDumper{Storage: repo}
}
