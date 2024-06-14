package utils

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
)

// DecodeBody decompresses a byte array (response body) based on the content encoding
func DecodeBody(body []byte, content_encoding string) ([]byte, error) {
	switch content_encoding {
	case "gzip":
		reader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("gzip decompress error: %v", err)
		}
		defer reader.Close()
		decompressedBody, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("gzip reader error: %v", err)
		}
		body = decompressedBody // Assign the decompressed body back to the original body variable
	case "deflate":
		reader := flate.NewReader(bytes.NewReader(body))
		defer reader.Close()
		decompressedBody, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("deflate reader error: %v", err)
		}
		body = decompressedBody // Assign the decompressed body back to the original body variable
	case "", "identity":
		// no encoding, do nothing
	default:
		return nil, fmt.Errorf("unsupported encoding: %s", content_encoding)
	}
	return body, nil
}
