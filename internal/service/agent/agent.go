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
		},
	}
}

func (a *Agent) Run() {
	var i = 1
	var updateToSendRatio = int(a.config.ReportInterval / a.config.PollInterval)
	for {
		a.poller.update()

		if i%updateToSendRatio == 0 {
			a.sender.send()
			i = 0
		}
		i++
		time.Sleep(a.config.PollInterval)
	}
}
