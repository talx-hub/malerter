package gzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/talx-hub/malerter/internal/constants"
)

type Writer struct {
	http.ResponseWriter
	compressor    *gzip.Writer
	wasCompressed bool
}

func NewWriter(w http.ResponseWriter) *Writer {
	return &Writer{
		ResponseWriter: w,
		compressor:     gzip.NewWriter(w),
		wasCompressed:  false,
	}
}

func needCompress(contentType string) bool {
	isHTML := strings.Contains(contentType, constants.ContentTypeHTML)
	isJSON := strings.Contains(contentType, constants.ContentTypeJSON)
	if isHTML || isJSON {
		return true
	}
	return false
}

func (w *Writer) Write(rawData []byte) (int, error) {
	contentType := w.ResponseWriter.Header().Get(constants.KeyContentType)
	if needCompress(contentType) {
		w.wasCompressed = true
		n, err := w.compressor.Write(rawData)
		if err != nil {
			return n, fmt.Errorf("unable to compress data: %w", err)
		}
		return n, nil
	}
	n, err := w.ResponseWriter.Write(rawData)
	if err != nil {
		return n, fmt.Errorf("unable to write response: %w", err)
	}
	return n, nil
}

func (w *Writer) WriteHeader(statusCode int) {
	contentType := w.ResponseWriter.Header().Get(constants.KeyContentType)
	if isOK(statusCode) && needCompress(contentType) {
		w.ResponseWriter.Header().Set(constants.KeyContentEncoding, "gzip")
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func isOK(statusCode int) bool {
	const codeFirstOK = 200
	const codeAfterLastOK = 300
	return statusCode >= codeFirstOK && statusCode < codeAfterLastOK
}

func (w *Writer) Close() error {
	if w.wasCompressed {
		err := w.compressor.Close()
		if err != nil {
			return fmt.Errorf("failed to close compressor: %w", err)
		}
	}
	return nil
}

func Middleware(h http.Handler) http.Handler {
	gzipFunc := func(w http.ResponseWriter, r *http.Request) {
		resultWriter := w
		acceptEncoding := r.Header.Get(constants.KeyAcceptEncoding)
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			compressor := NewWriter(w)
			resultWriter = compressor
			defer func() {
				if err := compressor.Close(); err != nil {
					log.Fatalf("unable to close compressing writer: %v", err)
				}
			}()
		}

		contentEncoding := r.Header.Get(constants.KeyContentEncoding)
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
					log.Fatalf("unable to close decompressing reader: %v", err)
				}
			}()
		}
		h.ServeHTTP(resultWriter, r)
	}

	return http.HandlerFunc(gzipFunc)
}

type Reader struct {
	io.ReadCloser
	decompressor *gzip.Reader
}

func NewReader(r io.ReadCloser) (*Reader, error) {
	decompressor, err := gzip.NewReader(r)
	if err != nil {
		return nil,
			fmt.Errorf("failed to construct reader for decompressor: %w", err)
	}

	return &Reader{
		ReadCloser:   r,
		decompressor: decompressor,
	}, nil
}

func (r *Reader) Read(compressedData []byte) (int, error) {
	n, err := r.decompressor.Read(compressedData)
	if err != nil {
		return n, fmt.Errorf("decompressor read error: %w", err)
	}
	return n, nil
}

func (r *Reader) Close() error {
	if err := r.ReadCloser.Close(); err != nil {
		return fmt.Errorf("failed to clore decompressor reader: %w", err)
	}
	err := r.decompressor.Close()
	if err != nil {
		return fmt.Errorf("failed to close decompressor: %w", err)
	}
	return nil
}

func Compress(data []byte) (*bytes.Buffer, error) {
	var compressed bytes.Buffer
	compressor := gzip.NewWriter(&compressed)
	_, err := compressor.Write(data)
	if err != nil {
		return nil,
			fmt.Errorf("failed write data to compress temporary buffer: %w", err)
	}
	err = compressor.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %w", err)
	}
	return &compressed, nil
}
