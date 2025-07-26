package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"errors"
	"fmt"
)

type Decrypter struct {
	PrivateKey *rsa.PrivateKey
	Nonce      [12]byte
}

func NewDecrypter(privateKeyPath string) (*Decrypter, error) {
	if privateKeyPath == "" {
		return nil, errors.New("no private key provided")
	}
	PrivateKey, err := GetPrivateKey(privateKeyPath)
	if err != nil {
		return nil, err
	}

	return &Decrypter{PrivateKey: PrivateKey, Nonce: [12]byte{}}, nil
}

func (d *Decrypter) Decrypt(encrypted []byte) ([]byte, error) {
	newKey, err := NewKey()
	if err != nil {
		return encrypted, err
	}
	if len(encrypted) < EncryptedSessionKeySize {
		return encrypted, errors.New("wrong key size")
	}
	err = newKey.Decrypt(d.PrivateKey, encrypted[:EncryptedSessionKeySize])
	if err != nil {
		return encrypted, fmt.Errorf("failed to decrypt: %w", err)
	}
	aesblock, err := aes.NewCipher(newKey)
	if err != nil {
		return encrypted, fmt.Errorf("failed to get aesblock: %w", err)
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return encrypted, fmt.Errorf("failed to get aesgcm: %w", err)
	}

	decrypted, err := aesgcm.Open(
		nil,
		d.Nonce[:], encrypted[EncryptedSessionKeySize:],
		nil)
	if err != nil {
		return decrypted, fmt.Errorf("failed to decrypt: %w", err)
	}

	return decrypted, nil
}
