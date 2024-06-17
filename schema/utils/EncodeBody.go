package utils

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"strconv"
	"strings"
)

const (
	gzipEncoding     = "gzip"
	deflateEncoding  = "deflate"
	identityEncoding = "identity"
)

// parseAcceptEncoding parses the Accept-Encoding header to find out what encodings are accepted.
func parseAcceptEncoding(headerValue *string) map[string]float64 {
	encodings := make(map[string]float64)
	if *headerValue == "" {
		// early return for empty header
		return encodings
	}

	for _, part := range strings.Split(*headerValue, ",") {
		pieces := strings.Split(strings.TrimSpace(part), ";q=")
		encoding := pieces[0]
		quality := 1.0 // default quality
		if len(pieces) > 1 {
			var err error
			quality, err = strconv.ParseFloat(pieces[1], 64)
			if err != nil {
				quality = 1.0 // in case of error, default to quality 1.0
			}
		}
		encodings[encoding] = quality
	}
	return encodings
}

// chooseEncoding selects the best encoding to use based on the client's Accept-Encoding header.
// This is a simplified example that prefers gzip over deflate.
func chooseEncoding(acceptEncodingHeader *string) string {
	encodings := parseAcceptEncoding(acceptEncodingHeader)

	// Example logic to choose encoding
	if quality, ok := encodings[gzipEncoding]; ok && quality > 0 {
		return gzipEncoding
	} else if quality, ok := encodings[deflateEncoding]; ok && quality > 0 {
		return deflateEncoding
	}
	return ""
}

// gzipCompress compresses a byte array using gzip
func gzipCompress(body []byte) ([]byte, string, error) {
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)

	_, err := writer.Write(body)
	if err != nil {
		writer.Close() // Attempt to close the writer even if there's an error
		return nil, "", fmt.Errorf("failed to write data for compression: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buffer.Bytes(), gzipEncoding, nil
}

// deflateCompress compresses a byte array using deflate
func deflateCompress(body []byte) ([]byte, string, error) {
	var buffer bytes.Buffer
	writer, err := flate.NewWriter(&buffer, flate.DefaultCompression)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create deflate writer: %w", err)
	}

	_, err = writer.Write(body)
	if err != nil {
		writer.Close() // Attempt to close the writer even if there's an error
		return nil, "", fmt.Errorf("failed to write data for compression: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to close deflate writer: %w", err)
	}

	return buffer.Bytes(), deflateEncoding, nil
}

// EncodeBody compresses a string (the response body) based on the content-encoding header
func EncodeBody(body *string, acceptEncodingHeader string) (encodedBody []byte, encoding string, err error) {
	selectedEncoding := chooseEncoding(&acceptEncodingHeader)
	bodyBytes := []byte(*body)

	switch selectedEncoding {
	case gzipEncoding:
		return gzipCompress(bodyBytes)
	case deflateEncoding:
		return deflateCompress(bodyBytes)
	case "", identityEncoding:
		return bodyBytes, selectedEncoding, nil
	}
	return nil, "", fmt.Errorf("unsupported encoding: %s", selectedEncoding)
}
