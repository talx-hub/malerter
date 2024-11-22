package zerologger

import (
	"fmt"
	"net/http"
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
		return nil, fmt.Errorf("unable to init logger: %v", err)
	}
	zerolog.DurationFieldUnit = time.Second
	logger := ZeroLogger{
		Logger: zerolog.New(zerolog.ConsoleWriter{
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
		w            http.ResponseWriter
		responseData *responseData
	}
)

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.w.Write(b)
	w.responseData.size += size
	return size, err
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.w.WriteHeader(statusCode)
	w.responseData.status = statusCode
}

func (w *loggingResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (logger ZeroLogger) Middleware(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		uri := r.RequestURI
		method := r.Method
		responseData := &responseData{}
		lw := loggingResponseWriter{
			w:            w,
			responseData: responseData,
		}

		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		logger.Info().
			Str("URI", uri).
			Str("method", method).
			Int("status", responseData.status).
			Int("size", responseData.size).
			Dur("duration", duration).
			Send()
	}

	return http.HandlerFunc(logFn)
}
