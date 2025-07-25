package middlewares

import (
	"bytes"
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/pkg/crypto"
)

func setupEncryption(t *testing.T) (enc *crypto.Encrypter, dec *crypto.Decrypter, cleanup func()) {
	privPath, pubPath := writeTestKeyPair(t)
	enc, err := crypto.NewEncrypter(pubPath)
	require.NoError(t, err)

	dec, err = crypto.NewDecrypter(privPath)
	require.NoError(t, err)

	cleanup = func() {
		_ = os.Remove(privPath)
		_ = os.Remove(pubPath)
	}
	return enc, dec, cleanup
}

func writeTestKeyPair(t *testing.T) (string, string) {
	privateKey, err := GenerateRSAKey()
	require.NoError(t, err)

	privDER := MarshalPrivateKeyBase64(privateKey)
	privFile, err := os.CreateTemp("", "priv-*.pem")
	require.NoError(t, err)
	_, _ = privFile.Write(privDER)
	_ = privFile.Close()

	pubDER := MarshalPublicKeyPEM(&privateKey.PublicKey)
	pubFile, err := os.CreateTemp("", "pub-*.pem")
	require.NoError(t, err)
	_, _ = pubFile.Write(pubDER)
	_ = pubFile.Close()

	return privFile.Name(), pubFile.Name()
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	data, _ := io.ReadAll(r.Body)
	_, _ = w.Write(data)
}

func TestDecryptMiddleware_UnencryptedRequest(t *testing.T) {
	_, dec, cleanup := setupEncryption(t)
	defer cleanup()

	log := logger.NewNopLogger()
	mw := Decrypt(dec, log)

	body := "plain text"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	resp := httptest.NewRecorder()

	mw(http.HandlerFunc(echoHandler)).ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, body, resp.Body.String())
}

func TestDecryptMiddleware_BadEncryptedBody(t *testing.T) {
	_, dec, cleanup := setupEncryption(t)
	defer cleanup()

	log := logger.NewNopLogger()
	mw := Decrypt(dec, log)

	bad := make([]byte, crypto.EncryptedSessionKeySize+20)
	_, _ = rand.Read(bad)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bad))
	req.Header.Set("X-Encrypted", "true")
	resp := httptest.NewRecorder()

	mw(http.HandlerFunc(echoHandler)).ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}
