package utils

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
)

// DecodeBody decompresses a byte array (response body) based on the content encoding
func DecodeBody(body []byte, content_encoding string) (decodedBody []byte, err error) {
	switch content_encoding {
	case "gzip":
		reader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("gzip decompress error: %v", err)
		}
		defer reader.Close()
		decodedBody, err = io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("gzip reader error: %v", err)
		}
	case "deflate":
		reader := flate.NewReader(bytes.NewReader(body))
		defer reader.Close()
		decodedBody, err = io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("deflate reader error: %v", err)
		}
	case "", "identity":
		// no encoding, do nothing
		return body, nil
	default:
		return nil, fmt.Errorf("unsupported encoding: %s", content_encoding)
	}

	return decodedBody, nil
}
