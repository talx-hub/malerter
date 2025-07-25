package crypto_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/talx-hub/malerter/pkg/crypto"
)

func generateRSAPublicKeyFile(t *testing.T) string {
	t.Helper()
	private, err := rsa.GenerateKey(rand.Reader, 1024)
	assert.NoError(t, err)
	pub := &private.PublicKey
	pubDER := x509.MarshalPKCS1PublicKey(pub)

	block := &pem.Block{
		Type:  crypto.PublicKeyTitle,
		Bytes: pubDER,
	}

	tmpFile, err := os.CreateTemp("", "pubkey-*.pem")
	assert.NoError(t, err)

	err = pem.Encode(tmpFile, block)
	assert.NoError(t, err)
	tmpFile.Close()

	return tmpFile.Name()
}

func TestNewEncrypter_Success(t *testing.T) {
	pubKeyFile := generateRSAPublicKeyFile(t)
	defer os.Remove(pubKeyFile)

	encrypter, err := crypto.NewEncrypter(pubKeyFile)
	assert.NoError(t, err)
	assert.NotNil(t, encrypter)
	assert.NotNil(t, encrypter.PublicKey)
}

func TestNewEncrypter_MissingPath(t *testing.T) {
	encrypter, err := crypto.NewEncrypter("")
	assert.Nil(t, encrypter)
	assert.EqualError(t, err, "no public key provided")
}

func TestNewEncrypter_InvalidFile(t *testing.T) {
	tmp := filepath.Join(os.TempDir(), "invalid.pem")
	os.WriteFile(tmp, []byte("not a key"), 0600)
	defer os.Remove(tmp)

	encrypter, err := crypto.NewEncrypter(tmp)
	assert.Nil(t, encrypter)
	assert.Error(t, err)
}

func TestEncrypter_Encrypt_Success(t *testing.T) {
	pubKeyFile := generateRSAPublicKeyFile(t)
	defer os.Remove(pubKeyFile)

	encrypter, err := crypto.NewEncrypter(pubKeyFile)
	assert.NoError(t, err)

	payload := []byte("secret-data")
	encrypted, err := encrypter.Encrypt(payload)
	assert.NoError(t, err)

	// Должен быть больше, чем просто payload
	assert.True(t, len(encrypted) > len(payload))
}
