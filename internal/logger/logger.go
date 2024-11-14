package logger

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	Logger zerolog.Logger
}

func New(logLevel string) (*Logger, error) {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return nil, fmt.Errorf("unable to init logger: %v", err)
	}
	zerolog.DurationFieldUnit = time.Second
	logger := Logger{
		zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.Stamp,
		}).
			Level(level).
			With().
			Timestamp().
			Int("pid", os.Getpid()).
			Logger(),
	}
	return &logger, nil
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.responseData.size += size
	return size, err
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.responseData.status = statusCode
}

func (logger Logger) WrapHandler(h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		uri := r.RequestURI
		method := r.Method
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		logger.Logger.Info().
			Str("URI", uri).
			Str("method", method).
			Int("status", responseData.status).
			Int("size", responseData.size).
			Dur("duration", duration).
			Send()
	}

	return logFn
}
