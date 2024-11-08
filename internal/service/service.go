package service

import (
	"github.com/talx-hub/malerter/internal/model"
)

type Service interface {
	Store(metric model.Metric) error
	Get(metric model.Metric) (model.Metric, error)
	GetAll() []model.Metric
}
