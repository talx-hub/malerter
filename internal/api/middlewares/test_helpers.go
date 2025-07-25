package middlewares

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/pkg/crypto"
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

func (h *sigStubHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write([]byte(testBody))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GenerateRSAKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

func MarshalPrivateKeyBase64(priv *rsa.PrivateKey) []byte {
	der := x509.MarshalPKCS1PrivateKey(priv)
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: der}
	pemData := pem.EncodeToMemory(block)
	pemData = bytes.TrimPrefix(pemData, []byte("-----BEGIN RSA PRIVATE KEY-----\n"))
	pemData = bytes.TrimSuffix(pemData, []byte("-----END RSA PRIVATE KEY-----\n"))
	return pemData
}

func MarshalPublicKeyPEM(pub *rsa.PublicKey) []byte {
	pubDER := x509.MarshalPKCS1PublicKey(pub)
	return pem.EncodeToMemory(&pem.Block{Type: crypto.PublicKeyTitle, Bytes: pubDER})
}
