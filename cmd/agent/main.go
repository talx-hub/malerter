package main

import (
	"context"
	"log"

	agentCfg "github.com/talx-hub/malerter/internal/config/agent"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/service/agent"
	"github.com/talx-hub/malerter/internal/service/agent/buildinfo"
	"github.com/talx-hub/malerter/pkg/shutdown"
)

func main() {
	cfg, ok := agentCfg.NewDirector().Build().(agentCfg.Builder)
	if !ok {
		log.Fatal("unable to load agent config")
	}
	zeroLogger, err := logger.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("unable to configure custom logger: %s", err.Error())
	}

	zeroLogger.Info().
		Str("buildVersion", buildinfo.Version).
		Str("buildCommit", buildinfo.Commit).
		Str("buildDate", buildinfo.Date).
		Msg("Starting agent")
	agt := agent.NewAgent(&cfg, zeroLogger)
	if agt == nil {
		zeroLogger.Fatal().Msg("agent failed. Exit")
		return
	}
	defer func() {
		err = agt.Close()
		if err != nil {
			zeroLogger.Fatal().Err(err).Msg("failed to close agent properly")
		}
	}()

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
