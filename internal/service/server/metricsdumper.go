package server

import (
	"context"

	"github.com/talx-hub/malerter/internal/model"
)

type Storage interface {
	Add(ctx context.Context, metric model.Metric) error
	Batch(context.Context, []model.Metric) error
	Find(ctx context.Context, key string) (model.Metric, error)
	Get(ctx context.Context) ([]model.Metric, error)
	Ping(context.Context) error
}

type MetricsDumper struct {
	Storage
}

func NewMetricsDumper(repo Storage) *MetricsDumper {
	return &MetricsDumper{Storage: repo}
}
