package db

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/service/server/logger"
	"github.com/talx-hub/malerter/internal/utils/pgcontainer"
	"github.com/talx-hub/malerter/internal/utils/queue"
)

const initTO = 5 * time.Second
const defaultTO = 3 * time.Second

var (
	getDSN func() string
	getDB  func() *DB
)

func runMain(m *testing.M, log *logger.ZeroLogger) (int, error) {
	container, err := initContainer(log)
	if err != nil {
		return 1, fmt.Errorf("failed to init DB: %w", err)
	}
	defer container.Close()

	db := getDB()
	defer db.Close()

	exitCode := m.Run()
	return exitCode, nil
}

func initContainer(log *logger.ZeroLogger) (*pgcontainer.PGContainer, error) {
	pg := pgcontainer.New(log)
	getDSN = func() string {
		return pg.GetDSN()
	}
	err := pg.RunContainer()
	if err != nil {
		return nil, fmt.Errorf("failed to run docker container: %w", err)
	}
	if err = initGetDB(log); err != nil {
		pg.Close()
		return nil, fmt.Errorf("failed to init test DB: %w", err)
	}
	return pg, nil
}

func initGetDB(log *logger.ZeroLogger) error {
	dsn := getDSN()

	initCtx, cancel := context.WithTimeout(context.Background(), initTO)
	defer cancel()

	buffer := queue.New[model.Metric]()
	defer buffer.Close()
	db, err := New(initCtx, dsn, log, &buffer)
	if err != nil {
		return fmt.Errorf("failed to init DB: %w", err)
	}

	getDB = func() *DB {
		return db
	}
	return nil
}

func benchmark[T any](
	b *testing.B,
	dbMethodAdapter func(context.Context, T) error,
	generator func() T,
	pg **pgcontainer.PGContainer,
) {
	b.Helper()

	log, err := logger.New(constants.LogLevelDefault)
	require.NoError(b, err)
	if pg != nil && *pg == nil {
		*pg, err = initContainer(log)
		require.NoError(b, err)
	}

	var ctx context.Context
	var cancel context.CancelFunc
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()

	for range b.N {
		ctx, cancel = context.WithTimeout(context.Background(), defaultTO)
		err = dbMethodAdapter(ctx, generator())
		require.NoError(b, err)
		cancel()
	}
}

func fillDB(b *testing.B, db *DB, container **pgcontainer.PGContainer) {
	b.Helper()

	adapterForAdd := func(ctx context.Context, val model.Metric) error {
		return db.Add(ctx, val)
	}
	benchmark(b, adapterForAdd, getMetricGenerator(), container)
}

func getMetricGenerator() func() model.Metric {
	i := 0
	return func() model.Metric {
		var m model.Metric
		if i%2 == 0 {
			value := float64(i)
			m = model.Metric{
				Value: &value,
				Type:  model.MetricTypeGauge,
				Name:  "m" + strconv.Itoa(i),
			}
		} else {
			delta := int64(i)
			m = model.Metric{
				Delta: &delta,
				Type:  model.MetricTypeCounter,
				Name:  "m" + strconv.Itoa(i),
			}
		}
		i++
		return m
	}
}

func getBatchGenerator() func() []model.Metric {
	return func() []model.Metric {
		prefix := "long-prefix-for-metric"
		delta := int64(1)
		return []model.Metric{
			{Name: prefix + "m1", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m2", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m3", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m4", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m5", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m6", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m7", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m8", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m9", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m10", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m11", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m12", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m13", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m14", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m15", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m16", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m17", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m18", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m19", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m20", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m21", Type: model.MetricTypeGauge, Value: new(float64)},
			{Name: prefix + "m22", Type: model.MetricTypeCounter, Delta: &delta},
		}
	}
}
