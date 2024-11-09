package repo

import (
	"github.com/talx-hub/malerter/internal/customerror"
	"github.com/talx-hub/malerter/internal/model"
)

type Repository interface {
	Store(metric model.Metric)
	Get(metric model.Metric) (model.Metric, error)
	GetAll() []model.Metric
}

type MemRepository struct {
	data map[string]model.Metric
}

func NewMemRepository() *MemRepository {
	return &MemRepository{data: make(map[string]model.Metric)}
}

func (r *MemRepository) Store(metric model.Metric) {
	dummyKey := metric.Type.String() + metric.Name
	if old, found := r.data[dummyKey]; found {
		err := old.Update(metric)
		if err != nil {
			return
		}
		n := old
		r.data[dummyKey] = n
	} else {
		r.data[dummyKey] = metric
	}
}

func (r *MemRepository) Get(metric model.Metric) (model.Metric, error) {
	dummyKey := metric.Type.String() + metric.Name
	if m, found := r.data[dummyKey]; found {
		return m, nil
	}
	return model.Metric{},
		&customerror.NotFoundError{MetricURL: metric.String()}
}

func (r *MemRepository) GetAll() []model.Metric {
	var metrics []model.Metric
	for _, m := range r.data {
		metrics = append(metrics, m)
	}
	return metrics
}
