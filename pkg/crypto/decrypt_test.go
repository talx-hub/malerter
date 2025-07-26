package crypto_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/talx-hub/malerter/pkg/crypto"
)

//nolint:unparam // для правильности
func writeTestKeyPair(t *testing.T) (string, string) {
	t.Helper()

	//nolint:gosec // для тестов достаточно такого размера ключа
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	assert.NoError(t, err)

	// ----- PRIVATE KEY -----
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// Note: ONLY base64, no headers; GetPrivateKey() adds them
	privPEMRaw := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privDER,
	})

	// Remove PEM headers manually
	privClean := bytes.TrimPrefix(privPEMRaw, []byte("-----BEGIN RSA PRIVATE KEY-----\n"))
	privClean = bytes.TrimSuffix(privClean, []byte("\n-----END RSA PRIVATE KEY-----\n"))

	privFile, err := os.CreateTemp("", "priv*.pem")
	assert.NoError(t, err)
	_, _ = privFile.Write(privClean)
	_ = privFile.Close()

	// ----- PUBLIC KEY -----
	pubDER := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  crypto.PublicKeyTitle,
		Bytes: pubDER,
	})

	pubFile, err := os.CreateTemp("", "pub*.pem")
	assert.NoError(t, err)
	_, _ = pubFile.Write(pubPEM)
	_ = pubFile.Close()

	return privFile.Name(), pubFile.Name()
}

func TestNewDecrypter_ValidKey(t *testing.T) {
	privPath, _ := writeTestKeyPair(t)
	defer func() {
		_ = os.Remove(privPath)
	}()

	d, err := crypto.NewDecrypter(privPath)
	assert.NoError(t, err)
	assert.NotNil(t, d)
	assert.NotNil(t, d.PrivateKey)
}

func TestNewDecrypter_EmptyPath(t *testing.T) {
	d, err := crypto.NewDecrypter("")
	assert.Nil(t, d)
	assert.EqualError(t, err, "no private key provided")
}

func TestDecrypter_Decrypt_ShortData(t *testing.T) {
	privPath, _ := writeTestKeyPair(t)
	defer func() {
		_ = os.Remove(privPath)
	}()

	decrypter, err := crypto.NewDecrypter(privPath)
	assert.NoError(t, err)

	encrypted := []byte("short")
	_, err = decrypter.Decrypt(encrypted)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wrong key size")
}

func TestDecrypter_Decrypt_CorruptedData(t *testing.T) {
	privPath, _ := writeTestKeyPair(t)
	defer func() {
		_ = os.Remove(privPath)
	}()

	decrypter, err := crypto.NewDecrypter(privPath)
	assert.NoError(t, err)

	// Валидная длина, но данные случайные
	corrupted := make([]byte, crypto.EncryptedSessionKeySize+32)
	_, _ = rand.Read(corrupted)

	_, err = decrypter.Decrypt(corrupted)
	assert.Error(t, err)
}
