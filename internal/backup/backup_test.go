package backup

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/memory"
)

const (
	backupFileName = "temp.bk"
)

func TestBackupRestore(t *testing.T) {
	cfg := server.Builder{
		FileStoragePath: backupFileName,
		StoreInterval:   3600,
	}

	log, err := logger.New(constants.LogLevelDefault)
	require.NoError(t, err)

	rep1 := memory.New(log)
	m1, _ := model.NewMetric().FromValues("mainQuestion", model.MetricTypeCounter, int64(42))
	m2, _ := model.NewMetric().FromValues("pi", model.MetricTypeGauge, 3.14)
	_ = rep1.Add(context.TODO(), m1)
	_ = rep1.Add(context.TODO(), m2)
	ms1, _ := rep1.Get(context.TODO())

	bk1 := New(&cfg, rep1, log)
	require.NotNil(t, bk1)
	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel1()
	bk1.backup(ctx1)

	rep2 := memory.New(log)
	bk2 := New(&cfg, rep2, log)
	require.NotNil(t, bk2)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel2()
	bk2.restore(ctx2)

	ms2, _ := rep2.Get(context.TODO())
	assert.ElementsMatch(t, ms1, ms2)
}

func TestRun(t *testing.T) {
	const (
		numberOfLoops = 2
		metricCount   = 2
	)

	const (
		fromPrevTest        = 2
		backupInLoop        = numberOfLoops * metricCount
		backupWhileShutdown = 2
	)

	cfg := server.Builder{
		FileStoragePath: backupFileName,
		StoreInterval:   3 * time.Second,
		Restore:         true,
	}

	log, err := logger.New(constants.LogLevelDefault)
	require.NoError(t, err)

	rep := memory.New(log)
	bk := New(&cfg, rep, log)
	require.NotNil(t, bk)

	timeout := numberOfLoops * cfg.StoreInterval
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	bk.Run(ctx)

	<-ctx.Done()
	r, err := newRestorer(backupFileName)
	require.NoError(t, err)
	defer func() {
		err := r.close()
		require.NoError(t, err)
	}()
	ms, err := r.read()
	require.NoError(t, err)

	assert.Equal(t, fromPrevTest+backupInLoop+backupWhileShutdown, len(ms))
}
