package repo

import (
	"github.com/talx-hub/malerter/internal/customerror"
)

type Repository interface {
	Store(metric Metric)
	Get(metric Metric) (Metric, error)
	GetAll() []Metric
}

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
	return Metric{},
		&customerror.NotFoundError{MetricURL: metric.String()}
}

func (r *MemRepository) GetAll() []Metric {
	var metrics []Metric
	for _, m := range r.data {
		metrics = append(metrics, m)
	}
	return metrics
}
