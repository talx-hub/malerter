package server

import (
	"context"

	"github.com/talx-hub/malerter/internal/model"
)

type Storage interface {
	Add(metric model.Metric) error
	Find(key string) (model.Metric, error)
	Get() []model.Metric
	Ping(context.Context) error
}

type MetricsDumper struct {
	Storage
}

func NewMetricsDumper(repo Storage) *MetricsDumper {
	return &MetricsDumper{Storage: repo}
}
