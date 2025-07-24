package backup

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/pkg/queue"
)

const (
	backupFileName = "temp.bk"
)

func TestBackupRestore(t *testing.T) {
	_ = os.Remove(backupFileName)

	cfg := server.Builder{
		FileStoragePath: backupFileName,
		StoreInterval:   3600,
	}

	log, err := logger.New(constants.LogLevelDefault)
	require.NoError(t, err)

	tunnel := queue.New[model.Metric]()
	rep1 := memory.New(log, &tunnel)
	m1, _ := model.NewMetric().FromValues("mainQuestion", model.MetricTypeCounter, int64(42))
	m2, _ := model.NewMetric().FromValues("pi", model.MetricTypeGauge, 3.14)
	_ = rep1.Add(context.TODO(), m1)
	_ = rep1.Add(context.TODO(), m2)
	ms1, _ := rep1.Get(context.TODO())

	bk1 := New(&cfg, &tunnel, rep1, log)
	require.NotNil(t, bk1)
	bk1.backup()

	rep2 := memory.New(log, nil)
	bk2 := New(&cfg, &tunnel, rep2, log)
	require.NotNil(t, bk2)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	bk2.restore(ctx)

	ms2, _ := rep2.Get(context.TODO())
	assert.ElementsMatch(t, ms1, ms2)
}

func TestRun(t *testing.T) {
	const (
		fromPrevTest  = 2
		backupInLoop  = 2
		numberOfLoops = 5
	)

	cfg := server.Builder{
		FileStoragePath: backupFileName,
		StoreInterval:   3 * time.Second,
		Restore:         true,
	}

	log, err := logger.New(constants.LogLevelDefault)
	require.NoError(t, err)

	tunnel := queue.New[model.Metric]()
	rep := memory.New(log, &tunnel)
	bk := New(&cfg, &tunnel, rep, log)
	require.NotNil(t, bk)

	timeout := numberOfLoops * cfg.StoreInterval
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	go bk.Run(ctx)

	m1, _ := model.NewMetric().FromValues("zero", model.MetricTypeCounter, int64(0))
	m2, _ := model.NewMetric().FromValues("e", model.MetricTypeGauge, 2.72)
	_ = rep.Add(context.TODO(), m1)
	time.Sleep(cfg.StoreInterval)
	time.Sleep(cfg.StoreInterval)
	_ = rep.Add(context.TODO(), m2)

	<-ctx.Done()
	r, err := newRestorer(backupFileName)
	require.NoError(t, err)
	defer func() {
		err := r.close()
		require.NoError(t, err)
	}()
	ms, err := r.read()
	require.NoError(t, err)

	assert.Equal(t, fromPrevTest+backupInLoop, len(ms))
}
