package repo

import (
	"github.com/talx-hub/malerter/internal/customerror"
	"github.com/talx-hub/malerter/internal/model"
)

type Repository interface {
	Store(metric model.Metric) error
	Get(key string) (model.Metric, error)
	GetAll() []model.Metric
}

type MemRepository struct {
	data map[string]model.Metric
}

func NewMemRepository() *MemRepository {
	return &MemRepository{data: make(map[string]model.Metric)}
}

func (r *MemRepository) Store(metric model.Metric) error {
	dummyKey := metric.Type.String() + metric.Name
	if old, found := r.data[dummyKey]; found {
		err := old.Update(metric)
		if err != nil {
			return err
		}
		n := old
		r.data[dummyKey] = n
	} else {
		r.data[dummyKey] = metric
	}
	return nil
}

func (r *MemRepository) Get(key string) (model.Metric, error) {
	if m, found := r.data[key]; found {
		return m, nil
	}
	return model.Metric{},
		&customerror.ErrNotFound{}
}

func (r *MemRepository) GetAll() []model.Metric {
	var metrics = make([]model.Metric, 0)
	for _, m := range r.data {
		metrics = append(metrics, m)
	}
	return metrics
}
