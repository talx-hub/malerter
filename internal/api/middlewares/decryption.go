package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/pkg/crypto"
)

func Decrypt(decryption *crypto.Decrypter, log *logger.ZeroLogger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		decrypt := func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Encrypted") != "true" {
				next.ServeHTTP(w, r)
				return
			}

			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				log.Warn().Err(err).Msg("failed to read body")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if err := r.Body.Close(); err != nil {
				log.Error().Err(err).Msg("failed to close body")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			data, err := decryption.Decrypt(bodyBytes)
			if err != nil {
				log.Error().Err(err).Msg("failed to decrypt body")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(data))

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(decrypt)
	}
}
