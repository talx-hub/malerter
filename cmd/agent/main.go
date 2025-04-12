package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	agentCfg "github.com/talx-hub/malerter/internal/config/agent"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/repository/memory"
	"github.com/talx-hub/malerter/internal/service/agent"
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
	rep := memory.New(zeroLogger, nil)
	agt := agent.NewAgent(rep, &cfg, &http.Client{}, zeroLogger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	idleAgentShutdown := make(chan struct{})
	go idleShutdown(idleAgentShutdown, zeroLogger, cancel)
	agt.Run(ctx)

	<-idleAgentShutdown
}

func idleShutdown(
	ch chan struct{},
	log *logger.ZeroLogger,
	cancelAgent context.CancelFunc,
) {
	defer close(ch)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	log.Info().Msg("shutdown signal received. Exiting...")
	cancelAgent()
}
