package service

import (
	"github.com/talx-hub/malerter/internal/model"
)

type Service interface {
	Add(metric model.Metric) error
	Find(key string) (model.Metric, error)
	Get() []model.Metric
}
