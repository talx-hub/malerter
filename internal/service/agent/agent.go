package agent

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/talx-hub/malerter/internal/config/agent"
	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/pkg/crypto"
)

type Sender interface {
	Send(ctx context.Context, jobs <-chan chan model.Metric, wg *sync.WaitGroup)
	Close() error
}

type Agent struct {
	config *agent.Builder
	poller Poller
	sender Sender
}

func NewAgent(
	cfg *agent.Builder,
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
	if cfg.UseGRPC {
		sender, err := NewGRPCSender(
			log,
			encrypter,
			cfg.ServerAddress,
			cfg.Secret,
		)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to start grpc agent")
			return nil
		}

		return &Agent{
			config: cfg,
			poller: Poller{
				log: log},
			sender: sender,
		}
	}

	return &Agent{
		config: cfg,
		poller: Poller{
			log: log},
		sender: &HTTPSender{
			host:      "http://" + cfg.ServerAddress,
			client:    &http.Client{},
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
	var wg sync.WaitGroup
	for {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return
		case <-pollTicker.C:
			temp := a.poller.update()
			jobs <- temp
		case <-reportTicker.C:
			for range a.config.RateLimit {
				wg.Add(1)
				go a.sender.Send(ctx, jobs, &wg)
			}
		}
	}
}

func (a *Agent) Close() error {
	err := a.sender.Close()
	if err != nil {
		return fmt.Errorf("failed to close agent: %w", err)
	}
	return nil
}

func makeJobsCh(cfg *agent.Builder) chan chan model.Metric {
	const safetyFactor = 2
	loopsCollected := int(cfg.ReportInterval / cfg.PollInterval)
	chanCap := safetyFactor * loopsCollected

	return make(chan chan model.Metric, chanCap)
}
