package repo

import (
	"github.com/alant1t/metricscoll/internal/customerror"
)

type Repository interface {
	Store(metric Metric) error
	Get(metric Metric) (Metric, error)
}

type MemRepository struct {
	data []Metric
}

func NewMemRepository() *MemRepository {
	return &MemRepository{}
}

func (r *MemRepository) Store(metric Metric) error {
	r.data = append(r.data, metric)
	return nil
}

func (r *MemRepository) Get(metric Metric) (Metric, error) {
	for _, e := range r.data {
		if e == metric {
			return e, nil
		}
	}
	// TODO: почему тут нужно возвращать адрес???
	return Metric{},
		&customerror.NotFoundError{RawMetric: metric.String()}
}
