package logger

import (
	"fmt"
	"io"
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

// NewNopLogger возвращает логгер, который игнорирует все записи.
// Полезен в тестах, когда логирование не требуется.
func NewNopLogger() *ZeroLogger {
	nopWriter := zerolog.New(io.Discard)
	return &ZeroLogger{Logger: nopWriter}
}
