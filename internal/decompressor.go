package internal

import (
	"bytes"
	"compress/gzip"
	"io"
)

// Decompress tries to decompress data using gzip.
// If decompression fails, returns the original data (assumes uncompressed).
func Decompress(data []byte) ([]byte, error) {
	// Try to decompress as gzip
	if r, err := gzip.NewReader(bytes.NewReader(data)); err == nil {
		defer r.Close()
		decompressed, err := io.ReadAll(r)
		if err == nil {
			return decompressed, nil
		}
	}

	// If decompression fails, assume data is uncompressed (valid per GELF spec)
	return data, nil
}
