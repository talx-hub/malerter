package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talx-hub/malerter/internal/repo"
)

func TestCollect(t *testing.T) {
	metrics := collect()
	t.Run("collect runtime metrics", func(t *testing.T) {
		require.Equal(t, len(metrics), MetricCount)
		for _, m := range metrics {
			assert.NoError(t, m.CheckValid())
		}
	})
}

func TestStore(t *testing.T) {
	storage := repo.NewMemRepository()
	poller := Poller{repo: storage}
	metrics := collect()
	poller.store(metrics)
	t.Run("store runtime metrics", func(t *testing.T) {
		stored := storage.GetAll()
		assert.Len(t, stored, MetricCount)
	})
}
