package repo

import (
	"github.com/alant1t/metricscoll/internal/customerror"
)

type Repository interface {
	Store(metric Metric)
	Get(metric Metric) (Metric, error)
	GetAll() []Metric
}

// TODO: подумать над тем, чтобы передать на массив
//   - если набор метрик фиксирован
type MemRepository struct {
	data map[string]Metric
}

func NewMemRepository() *MemRepository {
	return &MemRepository{data: make(map[string]Metric)}
}

func (r *MemRepository) Store(metric Metric) {
	dummyKey := metric.Type.String() + metric.Name
	if old, found := r.data[dummyKey]; found {
		old.Update(metric)
		n := old
		r.data[dummyKey] = n
	} else {
		r.data[dummyKey] = metric
	}
}

func (r *MemRepository) Get(metric Metric) (Metric, error) {
	dummyKey := metric.Type.String() + metric.Name
	if m, found := r.data[dummyKey]; found {
		return m, nil
	}
	// TODO: почему тут нужно возвращать адрес???
	return Metric{},
		&customerror.NotFoundError{RawMetric: metric.String()}
}

func (r *MemRepository) GetAll() []Metric {
	var metrics []Metric
	for _, m := range r.data {
		metrics = append(metrics, m)
	}
	return metrics
}
