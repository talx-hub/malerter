package middlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/talx-hub/malerter/internal/logger"
)

func setupTestLogger(t *testing.T) (*logger.ZeroLogger, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer

	log := logger.ZeroLogger{
		Logger: logger.NewNopLogger().Output(&buf),
	}
	return &log, &buf
}

func TestLoggingMiddleware_LogsCorrectly(t *testing.T) {
	log, buf := setupTestLogger(t)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("OK"))
	})

	loggedHandler := Logging(log)(testHandler)

	req := httptest.NewRequest(http.MethodPost, "/test-logging", http.NoBody)
	rec := httptest.NewRecorder()

	loggedHandler.ServeHTTP(rec, req)

	res := rec.Result()
	defer func() {
		_ = res.Body.Close()
	}()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "OK", rec.Body.String())

	logStr := buf.String()

	assert.Contains(t, logStr, `"method":"POST"`)
	assert.Contains(t, logStr, `"URI":"/test-logging"`)
	assert.Contains(t, logStr, `"status":201`)
	assert.Contains(t, logStr, `"size":2`) // "OK" = 2 bytes
	assert.Contains(t, logStr, `"duration":`)
}

func TestLoggingResponseWriter_WriteAndHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	respData := &responseData{}
	lrw := &loggingResponseWriter{
		w:            rec,
		responseData: respData,
	}

	header := lrw.Header()
	assert.NotNil(t, header)

	lrw.WriteHeader(http.StatusTeapot)
	assert.Equal(t, http.StatusTeapot, respData.status)

	n, err := lrw.Write([]byte("abc"))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, 3, respData.size)
}
