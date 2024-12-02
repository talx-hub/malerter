package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talx-hub/malerter/internal/repository/memory"
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
	storage := memory.New()
	poller := Poller{storage: storage}
	metrics := collect()
	poller.store(metrics)
	t.Run("store runtime metrics", func(t *testing.T) {
		stored := storage.Get()
		assert.Len(t, stored, MetricCount)
	})
}
