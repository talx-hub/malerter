package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/utils/compressor"
)

func TestCompress(t *testing.T) {
	stub := gzipStubHandler{}
	log, err := logger.New("Debug")
	require.NoError(t, err)
	withCompress := Compress(log)(&stub)
	srv := httptest.NewServer(withCompress)
	defer srv.Close()

	req, err := http.NewRequest(
		http.MethodPost, srv.URL, strings.NewReader(testBody))
	require.NoError(t, err)
	req.Header.Set(constants.KeyAcceptEncoding, "gzip")

	resp, _ := http.DefaultClient.Do(req)

	contentEncoding := resp.Header.Get(constants.KeyContentEncoding)
	require.True(t, strings.Contains(contentEncoding, "gzip"))

	var buf []byte
	decompressor, _ := gzip.NewReader(resp.Body)
	_, err = decompressor.Read(buf)
	d, err := io.ReadAll(decompressor)
	require.NoError(t, err)
	require.Equal(t, testBody, string(d))
	_ = resp.Body.Close()
}

func TestDecompress(t *testing.T) {
	stub := gzipStubHandler{}
	log, err := logger.New("Debug")
	require.NoError(t, err)
	withCompress := Decompress(log)(&stub)
	srv := httptest.NewServer(withCompress)
	defer srv.Close()

	buf, err := compressor.Compress([]byte(testBody))
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, srv.URL, buf)
	require.NoError(t, err)
	req.Header.Set(constants.KeyContentEncoding, "gzip")

	resp, _ := http.DefaultClient.Do(req)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	_ = resp.Body.Close()
	require.Equal(t, testBody, string(body))
}

func BenchmarkCompress(b *testing.B) {
	stub := gzipStubHandler{}
	log, err := logger.New("Debug")
	require.NoError(b, err)
	withCompress := Compress(log)(&stub)
	srv := httptest.NewServer(withCompress)
	defer srv.Close()
	b.ResetTimer()

	for range b.N {
		b.StopTimer()
		req, err := http.NewRequest(
			http.MethodPost, srv.URL, strings.NewReader(testBody))
		require.NoError(b, err)
		req.Header.Set(constants.KeyAcceptEncoding, "gzip")

		b.StartTimer()
		resp, _ := http.DefaultClient.Do(req)
		b.StopTimer()

		contentEncoding := resp.Header.Get(constants.KeyContentEncoding)
		require.True(b, strings.Contains(contentEncoding, "gzip"))

		var buf []byte
		decompressor, _ := gzip.NewReader(resp.Body)
		_, err = decompressor.Read(buf)
		d, err := io.ReadAll(decompressor)
		require.NoError(b, err)
		require.Equal(b, testBody, string(d))
		_ = resp.Body.Close()
	}
}

func BenchmarkDecompress(b *testing.B) {
	stub := gzipStubHandler{}
	log, err := logger.New("Debug")
	require.NoError(b, err)
	withCompress := Decompress(log)(&stub)
	srv := httptest.NewServer(withCompress)
	defer srv.Close()
	b.ResetTimer()

	for range b.N {
		b.StopTimer()
		buf, err := compressor.Compress([]byte(testBody))
		require.NoError(b, err)
		req, err := http.NewRequest(http.MethodPost, srv.URL, buf)
		require.NoError(b, err)
		req.Header.Set(constants.KeyContentEncoding, "gzip")

		b.StartTimer()
		resp, _ := http.DefaultClient.Do(req)
		b.StopTimer()

		body, err := io.ReadAll(resp.Body)
		require.NoError(b, err)
		_ = resp.Body.Close()
		require.Equal(b, testBody, string(body))
	}
}
