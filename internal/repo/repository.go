package repo

import (
	"github.com/alant1t/metricscoll/internal/customerror"
)

type Repository interface {
	Store(metric Metric) error
	Get(metric Metric) (Metric, error)
}

// TODO: подумать над тем, чтобы передать на массив
//   - если набор метрик фиксирован
type MemRepository struct {
	data map[string]Metric
}

func NewMemRepository() *MemRepository {
	return &MemRepository{data: make(map[string]Metric)}
}

func (r *MemRepository) Store(metric Metric) error {
	dummyKey := metric.Type + metric.Name
	if old, found := r.data[dummyKey]; found {
		old.Update(metric)
		n := old
		r.data[dummyKey] = n
	} else {
		r.data[dummyKey] = metric
	}
	return nil
}

func (r *MemRepository) Get(metric Metric) (Metric, error) {
	dummyKey := metric.Type + metric.Name
	if m, found := r.data[dummyKey]; found {
		return m, nil
	}
	// TODO: почему тут нужно возвращать адрес???
	return Metric{},
		&customerror.NotFoundError{RawMetric: metric.String()}
}
