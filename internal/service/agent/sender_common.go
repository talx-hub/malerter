package agent

import (
	"fmt"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/pkg/compressor"
	"github.com/talx-hub/malerter/pkg/crypto"
	"github.com/talx-hub/malerter/pkg/signature"
)

func (s *HTTPSender) tryCompress(data []byte) ([]byte, error) {
	if !s.compress {
		return data, nil
	}
	body, err := compressor.Compress(data)
	if err != nil {
		s.log.Error().Err(err).
			Msgf("unable to compress json %s", string(data))
		return nil, fmt.Errorf("failed to compress data: %w", err)
	}
	return body.Bytes(), nil
}

func trySign(data []byte, secret string) string {
	if secret != constants.NoSecret {
		return signature.Hash(data, secret)
	}
	return ""
}

func tryEncrypt(data []byte, encrypter *crypto.Encrypter) ([]byte, error) {
	if encrypter == nil {
		return data, nil
	}
	encryptedPayload, err := encrypter.Encrypt(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}
	return encryptedPayload, nil
}
