package agent

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/talx-hub/malerter/internal/config/agent"
	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/pkg/crypto"
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
	var encrypter *crypto.Encrypter
	if cfg.CryptoKeyPath != constants.EmptyPath {
		var err error
		encrypter, err = crypto.NewEncrypter(cfg.CryptoKeyPath)
		if err != nil {
			log.Error().Err(err).Msg("failed to add encryption to agent")
		}
	}

	return &Agent{
		config: cfg,
		poller: Poller{
			log: log},
		sender: Sender{
			host:      "http://" + cfg.ServerAddress,
			client:    client,
			compress:  true,
			log:       log,
			secret:    cfg.Secret,
			encrypter: encrypter,
		},
	}
}

func (a *Agent) Run(ctx context.Context) {
	pollTicker := time.NewTicker(a.config.PollInterval)
	reportTicker := time.NewTicker(a.config.ReportInterval)
	jobs := makeJobsCh(a.config)
	var m sync.Mutex
	var wg sync.WaitGroup
	for {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return
		case <-pollTicker.C:
			temp := a.poller.update()
			m.Lock()
			jobs <- temp
			m.Unlock()
		case <-reportTicker.C:
			for range a.config.RateLimit {
				wg.Add(1)
				go a.sender.send(jobs, &m, &wg)
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
