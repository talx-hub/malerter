package middlewares

import (
	"io"
	"net/http"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
)

const key = "very-secret-super-key"
const testBody = "some long very long body for signing in middleware"

type gzipStubHandler struct {
	log *logger.ZeroLogger
}

func (h *gzipStubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bodyToEcho, err := io.ReadAll(r.Body)
	defer func() {
		err = r.Body.Close()
		if err != nil {
			h.log.Error().Err(err).Msg("failed to close body")
		}
	}()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(constants.KeyContentType, constants.ContentTypeJSON)
	_, err = w.Write(bodyToEcho)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type sigStubHandler struct {
}

func (h *sigStubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(testBody))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
