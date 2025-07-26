package crypto

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

const (
	RandomSize              int = 2 * aes.BlockSize
	EncryptedSessionKeySize int = 512
)

type Key []byte

func NewKey() (Key, error) {
	key, err := GenerateRandom(RandomSize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random: %w", err)
	}

	return key, nil
}

func (k Key) Encrypt(public *rsa.PublicKey) ([]byte, error) {
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, public, k)
	if err != nil {
		return encrypted, fmt.Errorf("failed to encrypt: %w", err)
	}

	return encrypted, nil
}

func (k Key) Decrypt(private *rsa.PrivateKey, encrypted []byte) error {
	if err := rsa.DecryptPKCS1v15SessionKey(rand.Reader, private, encrypted, k); err != nil {
		return fmt.Errorf("failed to decrypt: %w", err)
	}

	return nil
}

func GenerateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return b, fmt.Errorf("failed to generate random: %w", err)
	}

	return b, nil
}

const (
	PublicKeyTitle = "RSA PUBLIC KEY"
)

func GetPrivateKey(fileName string) (*rsa.PrivateKey, error) {
	op := "GetPrivateKey"
	b, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("%s failed with an error: %w", op, err)
	}

	pemStr := string(b)

	pemStr = "-----BEGIN RSA PRIVATE KEY-----\n" + pemStr + "\n-----END RSA PRIVATE KEY-----"

	block, _ := pem.Decode([]byte(pemStr))

	if block == nil {
		return nil, errors.New(op + " failed to decode PEM block containing private key")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return key, nil
}

func GetPublicKey(publicKeyPath string) (*rsa.PublicKey, error) {
	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed with an error: %w", err)
	}
	block, _ := pem.Decode(publicKeyBytes)
	if block == nil || block.Type != PublicKeyTitle {
		return nil, errors.New("failed to decode PEM block containing public key")
	}
	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return key, nil
}
