// Package compressor предоставляет функцию сжатия данных с помощью алгоритма gzip.
//
// Используется для уменьшения размера передаваемых или сохраняемых данных.
//
// Пример использования:
//
//	compressed, err := compressor.Compress([]byte("your data here"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	io.Copy(dst, compressed)
package compressor

import (
	"bytes"
	"compress/gzip"
	"fmt"
)

// Compress принимает байтовый срез данных и возвращает буфер, содержащий сжатую gzip-версию этих данных.
//
// Возвращает:
//   - *bytes.Buffer с результатом сжатия;
//   - ошибку, если возникла ошибка при записи или закрытии gzip.Writer.
//
// Пример:
//
//	buf, err := compressor.Compress([]byte("example data"))
//	if err != nil {
//	    // обработка ошибки
//	}
//	io.Copy(os.Stdout, buf)
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
