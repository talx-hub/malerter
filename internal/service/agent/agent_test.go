package agent

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/talx-hub/malerter/internal/config/agent"
	"github.com/talx-hub/malerter/internal/service/server/logger"
)

func TestMakeJobsCh(t *testing.T) {
	cfg := &agent.Builder{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}
	ch := makeJobsCh(cfg)
	expectedCap := 2 * (int(cfg.ReportInterval / cfg.PollInterval))
	assert.Equal(t, expectedCap, cap(ch))
}

func TestNewAgent(t *testing.T) {
	cfg := &agent.Builder{
		ServerAddress:  "localhost:8080",
		PollInterval:   1 * time.Second,
		ReportInterval: 2 * time.Second,
		Secret:         "test-secret",
	}
	client := &http.Client{}
	log := logger.NewNopLogger()

	a := NewAgent(cfg, client, log)

	assert.Equal(t, cfg, a.config)
	assert.NotNil(t, a.poller)
	assert.NotNil(t, a.sender)
	assert.Equal(t, "http://localhost:8080", a.sender.host)
	assert.Equal(t, client, a.sender.client)
	assert.True(t, a.sender.compress)
	assert.Equal(t, "test-secret", a.sender.secret)
}
