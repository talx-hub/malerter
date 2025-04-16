package shutdown

import (
	"os"
	"os/signal"

	"github.com/talx-hub/malerter/internal/logger"
)

type CancelFunc func(args ...any) error

func IdleShutdown(
	idleCh chan struct{},
	log *logger.ZeroLogger,
	cancelFunc CancelFunc,
) {
	defer close(idleCh)

	sigintCh := make(chan os.Signal, 1)
	signal.Notify(sigintCh, os.Interrupt)
	<-sigintCh

	log.Info().Msg("shutdown signal received. Exiting....")
	if err := cancelFunc(); err != nil {
		log.Error().Err(err).Msg("error during service shutdown")
	}
}
