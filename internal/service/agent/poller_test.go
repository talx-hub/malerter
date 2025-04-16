package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
)

func TestRuntimeCollect(t *testing.T) {
	metrics := collectRuntime()
	t.Run("collect runtime metrics", func(t *testing.T) {
		assert.Equal(t, len(metrics), runtimeMetricCount)
		for _, m := range metrics {
			assert.NoError(t, m.CheckValid())
		}
	})
}

const psutilMinimumCount = 3

func TestPsutilCollect(t *testing.T) {
	metrics, err := collectPsutil()
	require.NoError(t, err)

	t.Run("collect psutil metrics", func(t *testing.T) {
		assert.GreaterOrEqual(t, len(metrics), psutilMinimumCount)
		for _, m := range metrics {
			assert.NoError(t, m.CheckValid())
		}
	})
}

func TestStore(t *testing.T) {
	log, err := logger.New(constants.LogLevelDefault)
	require.NoError(t, err)
	poller := Poller{log: log}
	t.Run("store runtime metrics", func(t *testing.T) {
		stored := poller.update()
		assert.GreaterOrEqual(t, len(stored), runtimeMetricCount+psutilMinimumCount)
	})
}
