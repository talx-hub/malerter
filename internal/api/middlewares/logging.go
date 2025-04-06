package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/talx-hub/malerter/internal/logger"
)

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
	if err != nil {
		return size,
			fmt.Errorf("failed to write with logging middleware %w", err)
	}
	w.responseData.size += size
	return size, nil
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.w.WriteHeader(statusCode)
	w.responseData.status = statusCode
}

func (w *loggingResponseWriter) Header() http.Header {
	return w.w.Header()
}

func Logging(log *logger.ZeroLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			responseData := &responseData{status: http.StatusOK}
			lw := loggingResponseWriter{
				w:            w,
				responseData: responseData,
			}

			next.ServeHTTP(&lw, r)
			log.Info().
				Str("URI", r.RequestURI).
				Str("method", r.Method).
				Int("status", responseData.status).
				Int("size", responseData.size).
				Dur("duration", time.Since(start)).
				Send()
		}
		return http.HandlerFunc(logFn)
	}
}
