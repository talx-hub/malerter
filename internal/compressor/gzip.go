package compressor

import (
	"bytes"
	"compress/gzip"
	"fmt"
)

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
