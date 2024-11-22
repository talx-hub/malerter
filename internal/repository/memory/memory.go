package memory

import (
	"sync"

	"github.com/talx-hub/malerter/internal/customerror"
	"github.com/talx-hub/malerter/internal/model"
)

type Metrics struct {
	data map[string]model.Metric
	m    *sync.RWMutex
}

func New() *Metrics {
	return &Metrics{data: make(map[string]model.Metric)}
}

func (r *Metrics) Add(metric model.Metric) error {
	dummyKey := metric.Type.String() + metric.Name

	r.m.Lock()
	defer r.m.Unlock()

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

func (r *Metrics) Find(key string) (model.Metric, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	if m, found := r.data[key]; found {
		return m, nil
	}
	return model.Metric{},
		&customerror.ErrNotFound{}
}

func (r *Metrics) Get() []model.Metric {
	r.m.RLock()
	defer r.m.RUnlock()

	var metrics = make([]model.Metric, 0)
	for _, m := range r.data {
		metrics = append(metrics, m)
	}
	return metrics
}
