package agent

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/internal/service/server/grpc_server"
)

func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrInt64(v int64) *int64 {
	return &v
}

func TestGRPCSender_Send(t *testing.T) {
	metrics := []model.Metric{
		{Name: "m1", Type: model.MetricTypeGauge},
		{Name: "m2", Type: model.MetricTypeGauge, Value: ptrFloat64(3.14)},
		{Name: "m3", Type: model.MetricTypeGauge, Value: ptrFloat64(3.15)},
		{Name: "m4", Type: model.MetricTypeGauge, Value: ptrFloat64(3.16)},
		{Name: "m5", Type: model.MetricTypeGauge, Value: ptrFloat64(3.17)},
		{Name: "m6", Type: model.MetricTypeCounter, Delta: ptrInt64(42)},
		{Name: "m6", Type: model.MetricTypeCounter, Delta: ptrInt64(42)},
		{Name: "m7", Type: model.MetricTypeCounter, Delta: ptrInt64(21)},
	}

	var j = make(chan model.Metric, len(metrics))
	for _, m := range metrics {
		j <- m
	}
	close(j)

	var jobs = make(chan chan model.Metric, 1)
	jobs <- j
	close(jobs)

	storage := memory.New(logger.NewNopLogger(), nil)
	srv := grpc_server.New(storage, logger.NewNopLogger())
	const addr = ":8081"
	go func() {
		lis, err := grpc_server.Serve(addr, srv)
		require.NoError(t, err)
		defer func() {
			err = lis.Close()
			require.NoError(t, err)
		}()
	}()

	time.Sleep(5 * time.Second)
	sender, err := NewGRPCSender(
		logger.NewNopLogger(),
		addr,
		"", false,
	)
	require.NoError(t, err)
	wg := sync.WaitGroup{}
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sender.Send(ctx, jobs, &wg)
	defer func() {
		err := sender.Close()
		require.NoError(t, err)
	}()
	time.Sleep(2 * time.Second)
	cancel()
	wg.Wait()

	result, err := storage.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, len(metrics)-2, len(result))
}
