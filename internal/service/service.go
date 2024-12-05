package service

import (
	"context"

	"github.com/talx-hub/malerter/internal/model"
)

type Service interface {
	Add(ctx context.Context, metric model.Metric) error
	Batch(context.Context, []model.Metric) error
	Find(ctx context.Context, key string) (model.Metric, error)
	Get(context.Context) ([]model.Metric, error)
	Ping(context.Context) error
}
