package memory

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/talx-hub/malerter/internal/customerror"
	"github.com/talx-hub/malerter/internal/model"
)

type Memory struct {
	data map[string]model.Metric
	m    sync.RWMutex
}

func New() *Memory {
	return &Memory{
		data: make(map[string]model.Metric),
	}
}

func (r *Memory) Add(_ context.Context, metric model.Metric) error {
	dummyKey := metric.Type.String() + " " + metric.Name

	r.m.Lock()
	defer r.m.Unlock()

	if old, found := r.data[dummyKey]; found {
		err := old.Update(metric)
		if err != nil {
			return fmt.Errorf("unable to update metric in storage: %w", err)
		}
		n := old
		r.data[dummyKey] = n
	} else {
		r.data[dummyKey] = metric
	}
	return nil
}

func (r *Memory) Batch(ctx context.Context, batch []model.Metric) error {
	for _, m := range batch {
		if err := r.Add(ctx, m); err != nil {
			log.Printf("failed to update batch metric: %v", err)
		}
	}
	return nil
}

func (r *Memory) Find(_ context.Context, key string) (model.Metric, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	if m, found := r.data[key]; found {
		return m, nil
	}
	return model.Metric{},
		&customerror.NotFoundError{}
}

func (r *Memory) Get(_ context.Context) ([]model.Metric, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	var metrics = make([]model.Metric, 0)
	for _, m := range r.data {
		metrics = append(metrics, m)
	}
	return metrics, nil
}

func (r *Memory) Ping(_ context.Context) error {
	return errors.New("a DB is not initialised, store in memory")
}

func (r *Memory) Clear() {
	r.data = make(map[string]model.Metric)
}
