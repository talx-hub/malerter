package service

import "github.com/talx-hub/malerter/internal/repo"

type Service interface {
	Store(string) error
	Get(string) (repo.Metric, error)
	GetAll() []repo.Metric
}
