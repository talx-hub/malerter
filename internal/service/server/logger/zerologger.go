package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type ZeroLogger struct {
	zerolog.Logger
}

func New(logLevel string) (*ZeroLogger, error) {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return nil, fmt.Errorf("unable to init logger: %w", err)
	}
	logger := ZeroLogger{
		Logger: zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).
			Level(level).
			With().
			Timestamp().
			Int("pid", os.Getpid()).
			Logger(),
	}
	return &logger, nil
}
