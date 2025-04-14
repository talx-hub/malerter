package agent

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/talx-hub/malerter/internal/config/agent"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
)

type Agent struct {
	config *agent.Builder
	poller Poller
	sender Sender
}

func NewAgent(
	cfg *agent.Builder,
	client *http.Client,
	log *logger.ZeroLogger,
) *Agent {

	return &Agent{
		config: cfg,
		poller: Poller{
			log: log},
		sender: Sender{
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
	jobs := makeJobsCh(a.config)
	var m sync.Mutex
	for {
		select {
		case <-ctx.Done():
			close(jobs)
			return

		case <-pollTicker.C:
			m.Lock()
			jobs <- a.poller.update()
			m.Unlock()

		case <-reportTicker.C:
			for range a.config.RateLimit {
				go a.sender.send(jobs, &m)
			}
		}
	}
}

func makeJobsCh(cfg *agent.Builder) chan chan model.Metric {
	const safetyFactor = 2
	loopsCollected := int(cfg.ReportInterval / cfg.PollInterval)
	chanCap := safetyFactor * loopsCollected

	return make(chan chan model.Metric, chanCap)
}
