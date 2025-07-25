package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewKey(t *testing.T) {
	key, err := NewKey()
	assert.NoError(t, err)
	assert.Len(t, key, RandomSize)
}

func TestGenerateRandom(t *testing.T) {
	b, err := GenerateRandom(16)
	assert.NoError(t, err)
	assert.Len(t, b, 16)
}

func TestKey_EncryptDecrypt(t *testing.T) {
	key, err := NewKey()
	assert.NoError(t, err)

	// Генерация RSA ключей
	private, err := rsa.GenerateKey(rand.Reader, 4096)
	assert.NoError(t, err)
	public := &private.PublicKey

	// Шифрование
	encrypted, err := key.Encrypt(public)
	assert.NoError(t, err)
	assert.Len(t, encrypted, EncryptedSessionKeySize)

	// Расшифровка
	err = key.Decrypt(private, encrypted)
	assert.NoError(t, err)
}

func writeTempPEMFile(t *testing.T, blockType string, derBytes []byte) string {
	t.Helper()
	file, err := os.CreateTemp("", "test-key-*.pem")
	assert.NoError(t, err)

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  blockType,
		Bytes: derBytes,
	})

	// удаляем "BEGIN ...", чтобы соответствовать `GetPrivateKey` логике
	if blockType == "RSA PRIVATE KEY" {
		stripped := bytes.TrimPrefix(pemBytes, []byte("-----BEGIN RSA PRIVATE KEY-----\n"))
		stripped = bytes.TrimSuffix(stripped, []byte("-----END RSA PRIVATE KEY-----\n"))
		_, err = file.Write(stripped)
	} else {
		_, err = file.Write(pemBytes)
	}
	assert.NoError(t, err)
	file.Close()

	return file.Name()
}

func TestGetPrivateKey_Success(t *testing.T) {
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	der := x509.MarshalPKCS1PrivateKey(private)
	path := writeTempPEMFile(t, "RSA PRIVATE KEY", der)
	defer os.Remove(path)

	loadedKey, err := GetPrivateKey(path)
	assert.NoError(t, err)
	assert.Equal(t, private.D, loadedKey.D)
}

func TestGetPublicKey_Success(t *testing.T) {
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	pubDER := x509.MarshalPKCS1PublicKey(&private.PublicKey)
	pubPath := writeTempPEMFile(t, PublicKeyTitle, pubDER)
	defer os.Remove(pubPath)

	pubKey, err := GetPublicKey(pubPath)
	assert.NoError(t, err)
	assert.Equal(t, private.PublicKey.N, pubKey.N)
}

func TestGetPrivateKey_InvalidPEM(t *testing.T) {
	path := filepath.Join(os.TempDir(), "invalid-priv.pem")
	_ = os.WriteFile(path, []byte("not a key"), 0600)
	defer os.Remove(path)

	_, err := GetPrivateKey(path)
	assert.Error(t, err)
}

func TestGetPublicKey_InvalidPEM(t *testing.T) {
	path := filepath.Join(os.TempDir(), "invalid-pub.pem")
	_ = os.WriteFile(path, []byte("not a key"), 0600)
	defer os.Remove(path)

	_, err := GetPublicKey(path)
	assert.Error(t, err)
}
