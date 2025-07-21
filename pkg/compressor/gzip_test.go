package compressor_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/talx-hub/malerter/pkg/compressor"
)

func TestCompress(t *testing.T) {
	input := []byte("test data to compress")

	compressed, err := compressor.Compress(input)
	if err != nil {
		t.Fatalf("Compress returned error: %v", err)
	}

	reader, err := gzip.NewReader(bytes.NewReader(compressed.Bytes()))
	if err != nil {
		t.Fatalf("gzip.NewReader error: %v", err)
	}
	defer func() {
		err := reader.Close()
		require.NoError(t, err)
	}()

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}

	if !bytes.Equal(input, output) {
		t.Errorf("decompressed data does not match original\nwant: %q\ngot:  %q", input, output)
	}
}
