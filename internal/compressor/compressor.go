package compressor

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Writer struct {
	http.ResponseWriter
	compressor *gzip.Writer
}

func NewWriter(w http.ResponseWriter) *Writer {
	return &Writer{
		ResponseWriter: w,
		compressor:     gzip.NewWriter(w),
	}
}

func (w *Writer) Write(rawData []byte) (int, error) {
	return w.compressor.Write(rawData)
}

func (w *Writer) WriteHeader(statusCode int) {
	if isOK(statusCode) {
		w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func isOK(statusCode int) bool {
	return statusCode < 300
}

func (w *Writer) Close() error {
	return w.compressor.Close()
}

type Reader struct {
	io.ReadCloser
	decompressor *gzip.Reader
}

func NewReader(r io.ReadCloser) (*Reader, error) {
	decompressor, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &Reader{
		ReadCloser:   r,
		decompressor: decompressor,
	}, nil
}

func (r *Reader) Read(compressedData []byte) (n int, err error) {
	return r.decompressor.Read(compressedData)
}

func (r *Reader) Close() error {
	if err := r.ReadCloser.Close(); err != nil {
		return err
	}
	return r.decompressor.Close()
}

func GzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resultWriter := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			compressor := NewWriter(w)
			resultWriter = compressor
			resultWriter.Header().Set("Content-Encoding", "gzip")
			defer func() {
				if err := compressor.Close(); err != nil {
					log.Fatal(err)
				}
			}()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			decompressor, err := NewReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = decompressor
			defer func() {
				if err = decompressor.Close(); err != nil {
					log.Fatal(err)
				}
			}()
		}
		h.ServeHTTP(resultWriter, r)
	}
}

func Compress(data []byte) (*bytes.Buffer, error) {
	var compressed bytes.Buffer
	compressor := gzip.NewWriter(&compressed)
	_, err := compressor.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}
	err = compressor.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}
	return &compressed, nil
}
