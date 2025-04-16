package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func Hash(data []byte, key string) string {
	s := hmac.New(sha256.New, []byte(key))
	s.Write(data)
	signature := s.Sum(nil)
	return hex.EncodeToString(signature)
}
