package service

import "github.com/talx-hub/malerter/internal/repo"

type Service interface {
	Store(metric repo.Metric) error
	Get(metric repo.Metric) (repo.Metric, error)
	GetAll() []repo.Metric
}
