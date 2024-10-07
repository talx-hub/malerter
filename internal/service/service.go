package service

import "github.com/alant1t/metricscoll/internal/repo"

type Service interface {
	Store(string) error
	Get(string) (repo.Metric, error)
	GetAll() []repo.Metric
}
