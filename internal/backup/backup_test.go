package backup

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/internal/repo"
)

func TestBackupRestore(t *testing.T) {
	m1, _ := model.NewMetric().FromValues("mainQuestion", model.MetricTypeCounter, int64(42))
	m2, _ := model.NewMetric().FromValues("pi", model.MetricTypeGauge, 3.14)
	rep1 := repo.NewMemRepository()
	_ = rep1.Store(m1)
	_ = rep1.Store(m2)
	cfg := server.Builder{FileStoragePath: "temp.bk"}

	bk1, err := New(cfg, rep1)
	require.NoError(t, err)
	defer func() {
		err1 := bk1.Close()
		require.NoError(t, err1)
	}()
	bk1.Backup()

	rep2 := repo.NewMemRepository()
	bk2, err := New(cfg, rep2)
	require.NoError(t, err)
	defer func() {
		err1 := bk2.Close()
		require.NoError(t, err1)
	}()
	bk2.Restore()

	assert.ElementsMatch(t, rep1.GetAll(), rep2.GetAll())
}
