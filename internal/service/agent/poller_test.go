package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/repository/memory"
)

func TestRuntimeCollect(t *testing.T) {
	metrics := collectRuntime()
	t.Run("collect runtime metrics", func(t *testing.T) {
		require.Equal(t, len(metrics), runtimeMetricCount)
		for _, m := range metrics {
			assert.NoError(t, m.CheckValid())
		}
	})
}

func TestPsutilCollect(t *testing.T) {
	metrics, err := collectPsutil()
	require.NoError(t, err)

	t.Run("collect psutil metrics", func(t *testing.T) {
		require.GreaterOrEqual(t, len(metrics), 3)
		for _, m := range metrics {
			assert.NoError(t, m.CheckValid())
		}
	})
}

func TestStore(t *testing.T) {
	log, err := logger.New(constants.LogLevelDefault)
	require.NoError(t, err)
	storage := memory.New(log, nil)
	poller := Poller{storage: storage, log: log}
	metrics := collectRuntime()
	poller.store(metrics)
	t.Run("store runtime metrics", func(t *testing.T) {
		stored, _ := storage.Get(context.TODO())
		assert.Len(t, stored, runtimeMetricCount)
	})
}
