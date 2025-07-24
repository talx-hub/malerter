package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/internal/customerror"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/memory"
)

func newMetric(name string, mType model.MetricType, value float64) model.Metric {
	switch mType {
	case model.MetricTypeCounter:
		delta := int64(value)
		return model.Metric{Name: name, Type: model.MetricTypeCounter, Delta: &delta}
	case model.MetricTypeGauge:
		return model.Metric{Name: name, Type: model.MetricTypeGauge, Value: &value}
	}
	return model.Metric{}
}

func TestMemory_AddAndFind(t *testing.T) {
	ctx := context.Background()
	log := logger.NewNopLogger()
	mem := memory.New(log, nil)

	metric := newMetric("CPU", model.MetricTypeGauge, 99.9)

	err := mem.Add(ctx, metric)
	require.NoError(t, err)

	found, err := mem.Find(ctx, "gauge CPU")
	require.NoError(t, err)
	assert.Equal(t, *metric.Value, *found.Value)
}

func TestMemory_Add_UpdateExisting(t *testing.T) {
	ctx := context.Background()
	log := logger.NewNopLogger()
	mem := memory.New(log, nil)

	metric := newMetric("RAM", model.MetricTypeGauge, 50.0)
	err := mem.Add(ctx, metric)
	require.NoError(t, err)

	updatedMetric := newMetric("RAM", model.MetricTypeGauge, 75.0)
	err = mem.Add(ctx, updatedMetric)
	require.NoError(t, err)

	found, err := mem.Find(ctx, "gauge RAM")
	require.NoError(t, err)
	assert.Equal(t, 75.0, *found.Value)
}

func TestMemory_Find_NotFound(t *testing.T) {
	ctx := context.Background()
	log := logger.NewNopLogger()
	mem := memory.New(log, nil)

	_, err := mem.Find(ctx, "gauge NotExist")
	assert.Error(t, err)
	var notFoundErr *customerror.NotFoundError
	assert.True(t, errors.As(err, &notFoundErr))
}

func TestMemory_Get(t *testing.T) {
	ctx := context.Background()
	log := logger.NewNopLogger()
	mem := memory.New(log, nil)

	metric1 := newMetric("M1", model.MetricTypeGauge, 10.2)
	metric2 := newMetric("M2", model.MetricTypeCounter, 20)

	_ = mem.Add(ctx, metric1)
	_ = mem.Add(ctx, metric2)

	all, err := mem.Get(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestMemory_Ping(t *testing.T) {
	log := logger.NewNopLogger()
	mem := memory.New(log, nil)

	err := mem.Ping(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "a DB is not initialised, store in memory", err.Error())
}

func TestMemory_Clear(t *testing.T) {
	ctx := context.Background()
	log := logger.NewNopLogger()
	mem := memory.New(log, nil)

	_ = mem.Add(ctx, newMetric("ClearMe", model.MetricTypeGauge, 1.1))
	mem.Clear()

	all, err := mem.Get(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 0)
}

func TestMemory_Batch(t *testing.T) {
	ctx := context.Background()
	log := logger.NewNopLogger()
	mem := memory.New(log, nil)

	batch := []model.Metric{
		newMetric("B1", model.MetricTypeGauge, 1.0),
		newMetric("B2", model.MetricTypeGauge, 2.0),
	}

	err := mem.Batch(ctx, batch)
	require.NoError(t, err)

	all, err := mem.Get(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 2)
}
