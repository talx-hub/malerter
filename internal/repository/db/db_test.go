package db

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/service/server/logger"
	"github.com/talx-hub/malerter/internal/utils/pgcontainer"
)

func TestMain(m *testing.M) {
	log, err := logger.New("DEBUG")
	if err != nil {
		os.Exit(-1)
	}
	code, err := runMain(m, log)
	if err != nil {
		log.Error().Err(err).Msg("unexpected test failure")
	}
	os.Exit(code)
}

func TestDB_Ping(t *testing.T) {
	db := getDB()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTO)
	defer cancel()
	err := db.Ping(ctx)
	require.NoError(t, err)

	cancel()
	err = db.Ping(ctx)
	require.Error(t, err)
}

func TestDB_Add(t *testing.T) {
	db := getDB()
	ctx, cancel := context.WithTimeout(context.Background(), defaultTO)
	defer cancel()
	m, err := model.NewMetric().FromValues("m1", model.MetricTypeGauge, 3.14)
	require.NoError(t, err)
	err = db.Add(ctx, m)
	require.NoError(t, err)
	err = db.Add(ctx, m)
	require.NoError(t, err)

	m, err = model.NewMetric().FromValues("m2", model.MetricTypeGauge, 2.71)
	require.NoError(t, err)
	err = db.Add(ctx, m)
	require.NoError(t, err)
	err = db.Add(ctx, m)
	require.NoError(t, err)
	m, err = model.NewMetric().FromValues("m2", model.MetricTypeGauge, 3.14)
	require.NoError(t, err)
	err = db.Add(ctx, m)
	require.NoError(t, err)

	m, err = model.NewMetric().FromValues("m1", model.MetricTypeCounter, int64(42))
	require.NoError(t, err)
	err = db.Add(ctx, m)
	require.NoError(t, err)
	err = db.Add(ctx, m)
	require.NoError(t, err)

	m, err = model.NewMetric().FromValues("m2", model.MetricTypeCounter, int64(0))
	require.NoError(t, err)
	err = db.Add(ctx, m)
	require.NoError(t, err)
	err = db.Add(ctx, m)
	require.NoError(t, err)

	m, err = model.NewMetric().FromValues("m3", model.MetricTypeCounter, int64(1))
	require.NoError(t, err)
	err = db.Add(ctx, m)
	require.NoError(t, err)
	err = db.Add(ctx, m)
	require.NoError(t, err)

	cancel()
	err = db.Add(ctx, m)
	require.Error(t, err)
}

func TestDB_Batch(t *testing.T) {
	metrics := []model.Metric{
		{Value: new(float64), Type: model.MetricTypeGauge, Name: "b1"},
		{Value: new(float64), Type: model.MetricTypeGauge, Name: "b2"},
		{Value: new(float64), Type: model.MetricTypeGauge, Name: "b3"},
		{Value: new(float64), Type: model.MetricTypeGauge, Name: "b4"},
		{Value: new(float64), Type: model.MetricTypeGauge, Name: "b5"},
		{Delta: new(int64), Type: model.MetricTypeCounter, Name: "b1"},
		{Delta: new(int64), Type: model.MetricTypeCounter, Name: "b2"},
		{Delta: new(int64), Type: model.MetricTypeCounter, Name: "b3"},
	}
	db := getDB()
	ctx, cancel := context.WithTimeout(context.Background(), defaultTO)
	defer cancel()
	err := db.Batch(ctx, metrics)
	require.NoError(t, err)

	metrics2 := make([]model.Metric, 0)
	err = db.Batch(ctx, metrics2)
	require.NoError(t, err)

	cancel()
	err = db.Batch(ctx, metrics)
	require.Error(t, err)
}

func TestDB_Find(t *testing.T) {
	db := getDB()
	ctx, cancel := context.WithTimeout(context.Background(), defaultTO)
	defer cancel()
	m1, err := db.Find(ctx, "gauge m1")
	require.NoError(t, err)
	assert.Equal(t, "m1", m1.Name)
	assert.Equal(t, model.MetricTypeGauge, m1.Type)
	assert.Equal(t, 3.14, *m1.Value)

	m2, err := db.Find(ctx, "gauge m2")
	require.NoError(t, err)
	assert.Equal(t, "m2", m2.Name)
	assert.Equal(t, model.MetricTypeGauge, m2.Type)
	assert.Equal(t, 3.14, *m2.Value)

	m1Counter, err := db.Find(ctx, "counter m1")
	require.NoError(t, err)
	assert.Equal(t, "m1", m1Counter.Name)
	assert.Equal(t, model.MetricTypeCounter, m1Counter.Type)
	assert.Equal(t, int64(84), *m1Counter.Delta)

	m2Counter, err := db.Find(ctx, "counter m2")
	require.NoError(t, err)
	assert.Equal(t, "m2", m2Counter.Name)
	assert.Equal(t, model.MetricTypeCounter, m2Counter.Type)
	assert.Equal(t, int64(0), *m2Counter.Delta)

	m3Counter, err := db.Find(ctx, "counter m3")
	require.NoError(t, err)
	assert.Equal(t, "m3", m3Counter.Name)
	assert.Equal(t, model.MetricTypeCounter, m3Counter.Type)
	assert.Equal(t, int64(2), *m3Counter.Delta)

	b1, err := db.Find(ctx, "gauge b1")
	require.NoError(t, err)
	assert.Equal(t, "b1", b1.Name)
	assert.Equal(t, model.MetricTypeGauge, b1.Type)
	assert.Equal(t, float64(0), *b1.Value)

	b1C, err := db.Find(ctx, "counter b1")
	require.NoError(t, err)
	assert.Equal(t, "b1", b1C.Name)
	assert.Equal(t, model.MetricTypeCounter, b1C.Type)
	assert.Equal(t, int64(0), *b1C.Delta)

	_, err = db.Find(ctx, "counter absent")
	require.Error(t, err)

	cancel()
	_, err = db.Find(ctx, "gauge b1")
	require.Error(t, err)
}

func TestDB_Get(t *testing.T) {
	db := getDB()
	ctx, cancel := context.WithTimeout(context.Background(), defaultTO)
	defer cancel()
	all, err := db.Get(ctx)
	require.NoError(t, err)
	assert.Equal(t, 13, len(all))

	cancel()
	_, err = db.Get(ctx)
	require.Error(t, err)
}

func TestDB_Close(t *testing.T) {
	db := getDB()
	ctx, cancel := context.WithTimeout(context.Background(), defaultTO)
	defer cancel()
	err := db.Ping(ctx)
	require.NoError(t, err)
	db.Close()
	err = db.Ping(ctx)
	fmt.Println(err)
	require.Error(t, err)
}

func BenchmarkDB_Add_only_new_metrics(b *testing.B) {
	db := getDB()
	defer db.Close()

	var pg *pgcontainer.PGContainer
	defer func() {
		if pg != nil {
			pg.Close()
		}
	}()
	b.StopTimer()
	b.ResetTimer()

	adapter := func(ctx context.Context, m model.Metric) error {
		b.StartTimer()
		err := db.Add(ctx, m)
		b.StopTimer()
		return err
	}
	benchmark(b, adapter, getMetricGenerator(), &pg)
}

func BenchmarkDB_Add_update_metrics(b *testing.B) {
	db := getDB()
	defer db.Close()

	var pg *pgcontainer.PGContainer
	defer func() {
		if pg != nil {
			pg.Close()
		}
	}()

	fillDB(b, db, &pg)
	b.StopTimer()
	b.ResetTimer()

	adapter := func(ctx context.Context, m model.Metric) error {
		b.StartTimer()
		err := db.Add(ctx, m)
		b.StopTimer()
		return err
	}
	benchmark(b, adapter, getMetricGenerator(), &pg)
}

func BenchmarkDB_Batch(b *testing.B) {
	db := getDB()
	defer db.Close()

	var pg *pgcontainer.PGContainer
	defer func() {
		if pg != nil {
			pg.Close()
		}
	}()
	b.StopTimer()
	b.ResetTimer()

	adapter := func(ctx context.Context, batch []model.Metric) error {
		b.StartTimer()
		err := db.Batch(ctx, batch)
		b.StopTimer()
		return err
	}
	benchmark(b, adapter, getBatchGenerator(), &pg)
}

func BenchmarkDB_Find(b *testing.B) {
	db := getDB()
	defer db.Close()

	var pg *pgcontainer.PGContainer
	defer func() {
		if pg != nil {
			pg.Close()
		}
	}()
	fillDB(b, db, &pg)

	freshGenerate := getMetricGenerator()
	adapter := func(ctx context.Context, m model.Metric) error {
		metric := freshGenerate()
		key := metric.Type.String() + " " + metric.Name

		b.StartTimer()
		_, err := db.Find(ctx, key)
		b.StopTimer()
		return err
	}
	b.StopTimer()
	b.ResetTimer()

	benchmark(b, adapter, getMetricGenerator(), &pg)
}

func BenchmarkDB_Get(b *testing.B) {
	db := getDB()
	defer db.Close()

	var pg *pgcontainer.PGContainer
	defer func() {
		if pg != nil {
			pg.Close()
		}
	}()
	fillDB(b, db, &pg)

	adapter := func(ctx context.Context, m model.Metric) error {
		b.StartTimer()
		_, err := db.Get(ctx)
		b.StopTimer()
		return err
	}
	b.StopTimer()
	b.ResetTimer()

	benchmark(b, adapter, getMetricGenerator(), &pg)
}
