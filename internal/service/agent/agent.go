package agent

import (
	"context"
	"net/http"
	"time"

	"github.com/talx-hub/malerter/internal/config/agent"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
)

type Storage interface {
	Add(context.Context, model.Metric) error
	Get(context.Context) ([]model.Metric, error)
	Clear()
}

type Agent struct {
	config  *agent.Builder
	storage Storage
	poller  Poller
	sender  Sender
}

func NewAgent(
	storage Storage,
	cfg *agent.Builder,
	client *http.Client,
	log *logger.ZeroLogger,
) *Agent {
	return &Agent{
		config:  cfg,
		storage: storage,
		poller:  Poller{storage: storage, log: log},
		sender: Sender{
			storage:  storage,
			host:     "http://" + cfg.ServerAddress,
			client:   client,
			compress: true,
			log:      log,
			secret:   cfg.Secret,
		},
	}
}

func (a *Agent) Run(ctx context.Context) {
	pollTicker := time.NewTicker(a.config.PollInterval)
	reportTicker := time.NewTicker(a.config.ReportInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			a.poller.update()
		case <-reportTicker.C:
			a.sender.send()
		}
	}
}
