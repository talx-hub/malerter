package middlewares

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/utils/signature"
)

func TestWriteSignature(t *testing.T) {
	stub := sigStubHandler{}
	withSignature := WriteSignature(key)(&stub)
	srv := httptest.NewServer(withSignature)
	defer srv.Close()

	req, err := http.NewRequest(
		http.MethodPost, srv.URL, strings.NewReader(testBody))
	require.NoError(t, err)

	resp, _ := http.DefaultClient.Do(req)
	gotSign := resp.Header.Get(constants.KeyHashSHA256)
	require.True(t, len(gotSign) != 0)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = resp.Body.Close()
	require.NoError(t, err)

	calcSign := signature.Hash(body, key)
	assert.Equal(t, calcSign, gotSign)
}

func TestCheckSignature(t *testing.T) {
	stub := sigStubHandler{}
	withSignature := CheckSignature(key)(&stub)
	srv := httptest.NewServer(withSignature)
	defer srv.Close()

	req, err := http.NewRequest(
		http.MethodPost, srv.URL, strings.NewReader(testBody))
	require.NoError(t, err)
	sign := signature.Hash([]byte(testBody), key)
	req.Header.Set(constants.KeyHashSHA256, sign)

	resp, _ := http.DefaultClient.Do(req)
	err = resp.Body.Close()
	require.NoError(t, err)
}

func BenchmarkCheckSignature(b *testing.B) {
	stub := sigStubHandler{}
	withSignature := CheckSignature(key)(&stub)
	srv := httptest.NewServer(withSignature)
	defer srv.Close()
	b.ResetTimer()

	for range b.N {
		b.StopTimer()
		req, err := http.NewRequest(
			http.MethodPost, srv.URL, strings.NewReader(testBody))
		require.NoError(b, err)
		sign := signature.Hash([]byte(testBody), key)
		req.Header.Set(constants.KeyHashSHA256, sign)

		b.StartTimer()
		resp, _ := http.DefaultClient.Do(req)
		b.StopTimer()
		_ = resp.Body.Close()
	}
}
