// Package signature предоставляет функцию генерации HMAC-SHA256 подписи.
//
// Используется для создания или проверки цифровых подписей,
// например, для проверки целостности и подлинности данных.
//
// Пример использования:
//
//	signature := signature.Hash([]byte("my message"), "my_secret_key")
//	fmt.Println("HMAC:", signature)
package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Hash возвращает HMAC-SHA256 подпись для заданных данных и секретного ключа.
//
// Аргументы:
//   - data: входные данные, подлежащие подписи;
//   - key: секретный ключ.
//
// Возвращает строку с шестнадцатеричным представлением HMAC-подписи.
//
// Пример:
//
//	signature := signature.Hash([]byte("important data"), "mysecret")
//	fmt.Println(signature)
func Hash(data []byte, key string) string {
	s := hmac.New(sha256.New, []byte(key))
	s.Write(data)
	signature := s.Sum(nil)
	return hex.EncodeToString(signature)
}
