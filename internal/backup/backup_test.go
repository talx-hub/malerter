package backup

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/memory"
)

func TestBackupRestore(t *testing.T) {
	m1, _ := model.NewMetric().FromValues("mainQuestion", model.MetricTypeCounter, int64(42))
	m2, _ := model.NewMetric().FromValues("pi", model.MetricTypeGauge, 3.14)
	log, err := logger.New(constants.LogLevelDefault)
	require.NoError(t, err)
	rep1 := memory.New(log)
	_ = rep1.Add(context.TODO(), m1)
	_ = rep1.Add(context.TODO(), m2)
	cfg := server.Builder{FileStoragePath: "temp.bk"}

	bk1 := New(&cfg, rep1, log)
	require.NotNil(t, bk1)
	defer bk1.Close()
	bk1.Backup()

	rep2 := memory.New(log)
	bk2 := New(&cfg, rep2, log)
	require.NotNil(t, bk2)
	defer bk2.Close()
	bk2.Restore()
	ms1, _ := rep1.Get(context.TODO())
	ms2, _ := rep2.Get(context.TODO())

	assert.ElementsMatch(t, ms1, ms2)
}
