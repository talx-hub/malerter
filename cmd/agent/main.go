package main

import (
	"context"
	"log"
	"net/http"

	agentCfg "github.com/talx-hub/malerter/internal/config/agent"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/service/agent"
	"github.com/talx-hub/malerter/internal/utils/shutdown"
)

// TODO: сделать клиент модульным:
//		+ модуль сбора метрик
//		+ модуль отправки метрик
//		+ первый модуль собирает метрики
//		? затем оповещает модуль отпраки, что данные готовы
//		? но как сделать нотификацию???

func main() {
	// TODO: тут какие-то кошмары с указателями(см. config/agent/builder/.Build())
	cfg, ok := agentCfg.NewDirector().Build().(agentCfg.Builder)
	if !ok {
		log.Fatal("unable to load agent config")
	}
	zeroLogger, err := logger.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("unable to configure custom logger: %s", err.Error())
	}

	zeroLogger.Info().Msg("start agent")
	agt := agent.NewAgent(&cfg, &http.Client{}, zeroLogger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	idleCh := make(chan struct{})
	go shutdown.IdleShutdown(
		idleCh,
		zeroLogger,
		func(args ...any) error {
			return shutdownAgent(cancel)
		},
	)
	agt.Run(ctx)

	<-idleCh
}

func shutdownAgent(cancelAgent context.CancelFunc) error {
	cancelAgent()
	return nil
}
