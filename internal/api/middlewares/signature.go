package middlewares

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/utils/signature"
)

type SigningWriter struct {
	http.ResponseWriter
	key string
	sig string
}

func NewSigningWriter(w http.ResponseWriter, key string) *SigningWriter {
	return &SigningWriter{
		ResponseWriter: w,
		key:            key,
	}
}

func (w *SigningWriter) Write(body []byte) (int, error) {
	if w.key != constants.NoSecret {
		sig := signature.Hash(body, w.key)
		w.ResponseWriter.Header().Set(constants.KeyHashSHA256, sig)
	}

	n, err := w.ResponseWriter.Write(body)
	if err != nil {
		return n, fmt.Errorf("unable to write response: %w", err)
	}
	return n, nil
}

func CheckSignature(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		var checkFunc func(w http.ResponseWriter, r *http.Request)

		if key == constants.NoSecret {
			checkFunc = func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			}
			return http.HandlerFunc(checkFunc)
		}

		checkFunc = func(w http.ResponseWriter, r *http.Request) {
			body, err := checkSignature(key, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			signingWriter := NewSigningWriter(w, key)
			next.ServeHTTP(signingWriter, r)
		}
		return http.HandlerFunc(checkFunc)
	}
}

func checkSignature(key string, r *http.Request) ([]byte, error) {
	body, err := getBody(r)
	if err != nil {
		return nil, fmt.Errorf("body error: %w", err)
	}
	if len(body) == 0 {
		fmt.Println("no body")
		return body, nil
	}

	sig := r.Header.Get(constants.KeyHashSHA256)
	if sig == "" {
		fmt.Println("no sig")
		return nil, errors.New("no signature detected")
	}

	fmt.Println("got sign:", sig)
	hash := signature.Hash(body, key)
	fmt.Println("hash:", hash)
	if hash != sig {
		return nil, errors.New("wrong signature detected")
	}

	return body, nil
}

func getBody(r *http.Request) (b []byte, err error) {
	b, err = io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("body read error: %w", err)
	}
	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("body close error: %w", closeErr))
		}
	}()
	return b, nil
}
