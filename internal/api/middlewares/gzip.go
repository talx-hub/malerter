package middlewares

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
)

type GzipWriter struct {
	http.ResponseWriter
	compressor    *gzip.Writer
	wasCompressed bool
}

func NewGzipWriter(w http.ResponseWriter) *GzipWriter {
	return &GzipWriter{
		ResponseWriter: w,
		compressor:     gzip.NewWriter(w),
		wasCompressed:  false,
	}
}

func needCompress(contentType string) bool {
	isHTML := strings.Contains(contentType, constants.ContentTypeHTML)
	isJSON := strings.Contains(contentType, constants.ContentTypeJSON)
	return isHTML || isJSON
}

func (w *GzipWriter) Write(rawData []byte) (int, error) {
	contentType := w.ResponseWriter.Header().Get(constants.KeyContentType)
	if needCompress(contentType) {
		w.wasCompressed = true
		w.ResponseWriter.Header().Set(constants.KeyContentEncoding, "gzip")
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

func (w *GzipWriter) Close() error {
	if w.wasCompressed {
		err := w.compressor.Close()
		if err != nil {
			return fmt.Errorf("failed to close compressor: %w", err)
		}
	}
	return nil
}

func Gzip(logg *logger.ZeroLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		gzipFunc := func(w http.ResponseWriter, r *http.Request) {
			resultWriter := w
			acceptEncoding := r.Header.Get(constants.KeyAcceptEncoding)
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				compressor := NewGzipWriter(w)
				resultWriter = compressor
				defer func() {
					if err := compressor.Close(); err != nil {
						logg.Error().
							Err(err).Msg("unable to close compressing writer")
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
						logg.Error().
							Err(err).Msg("unable to close decompressing reader")
					}
				}()
			}
			next.ServeHTTP(resultWriter, r)
		}

		return http.HandlerFunc(gzipFunc)
	}
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

func (r *Reader) Read(dstDecompressed []byte) (int, error) {
	n, err := r.decompressor.Read(dstDecompressed)
	if err != nil {
		// странное! Если не добавить эту проверку и не возвращать io.EOF не обернутый,
		// то зацикливаемся когда где-то будем вызывать io.ReadAll
		if errors.Is(err, io.EOF) {
			return n, io.EOF
		}

		return n, fmt.Errorf("decompressor read error: %w", err)
	}
	return n, nil
}

func (r *Reader) Close() error {
	if err := r.ReadCloser.Close(); err != nil {
		return fmt.Errorf("failed to close decompressor reader: %w", err)
	}
	if err := r.decompressor.Close(); err != nil {
		return fmt.Errorf("failed to close decompressor: %w", err)
	}
	return nil
}
