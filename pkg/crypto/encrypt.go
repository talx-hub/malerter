package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"errors"
	"fmt"
)

type Encrypter struct {
	PublicKey *rsa.PublicKey
	Nonce     [12]byte
}

func NewEncrypter(publicKeyPath string) (*Encrypter, error) {
	if publicKeyPath == "" {
		return nil, errors.New("no public key provided")
	}
	PublicKey, err := GetPublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	return &Encrypter{PublicKey: PublicKey, Nonce: [12]byte{}}, nil
}

func (e *Encrypter) Encrypt(payload []byte) ([]byte, error) {
	key, err := NewKey()
	if err != nil {
		return payload, fmt.Errorf("failed to get new key: %w", err)
	}
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return payload, fmt.Errorf("failed to get aes block: %w", err)
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return payload, fmt.Errorf("failed to get aesgcm: %w", err)
	}
	eKey, err := key.Encrypt(e.PublicKey)
	if err != nil {
		return payload, fmt.Errorf("failed to encrypt: %w", err)
	}
	encrypted := make([]byte, 0, len(payload)+EncryptedSessionKeySize+len(key))
	encrypted = append(encrypted, eKey...)
	encrypted = append(encrypted, aesgcm.Seal(nil, e.Nonce[:], payload, nil)...)

	return encrypted, nil
}
