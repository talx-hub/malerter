package repo

import (
	"github.com/alant1t/metricscoll/internal/customerror"
	"github.com/alant1t/metricscoll/internal/service"
)

type Repository interface {
	Store(metric service.Metric) error
	Get(metric service.Metric) (string, error)
}

type MemRepository struct {
	data []service.Metric
}

func NewMemRepository() *MemRepository {
	return &MemRepository{}
}

func (r *MemRepository) Store(metric string) error {
	r.data = append(r.data, service.Metric{})
	return nil
}

func (r *MemRepository) Get(metric service.Metric) (service.Metric, error) {
	for _, e := range r.data {
		if e == metric {
			return e, nil
		}
	}
	// TODO: почему тут нужно возвращать адрес???
	return service.Metric{},
		&customerror.NotFoundError{RawMetric: metric.String()}
}
