package backup

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repository/memory"
)

func TestBackupRestore(t *testing.T) {
	m1, _ := model.NewMetric().FromValues("mainQuestion", model.MetricTypeCounter, int64(42))
	m2, _ := model.NewMetric().FromValues("pi", model.MetricTypeGauge, 3.14)
	rep1 := memory.New()
	_ = rep1.Add(context.TODO(), m1)
	_ = rep1.Add(context.TODO(), m2)
	cfg := server.Builder{FileStoragePath: "temp.bk"}

	bk1, err := New(&cfg, rep1)
	require.NoError(t, err)
	defer func() {
		err1 := bk1.Close()
		require.NoError(t, err1)
	}()
	bk1.Backup()

	rep2 := memory.New()
	bk2, err := New(&cfg, rep2)
	require.NoError(t, err)
	defer func() {
		err1 := bk2.Close()
		require.NoError(t, err1)
	}()
	bk2.Restore()
	ms1, _ := rep1.Get(context.TODO())
	ms2, _ := rep2.Get(context.TODO())

	assert.ElementsMatch(t, ms1, ms2)
}
