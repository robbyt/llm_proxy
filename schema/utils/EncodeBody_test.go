package utils

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAcceptEncoding(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected map[string]float64
	}{
		{
			name:   "Single encoding without quality",
			header: "gzip",
			expected: map[string]float64{
				"gzip": 1.0,
			},
		},
		{
			name:   "Multiple encodings with and without quality",
			header: "gzip, deflate;q=0.5, br;q=0",
			expected: map[string]float64{
				"gzip":    1.0,
				"deflate": 0.5,
				"br":      0.0,
			},
		},
		{
			name:     "Empty Accept-Encoding header",
			header:   "",
			expected: map[string]float64{},
		},
		{
			name:   "Invalid quality value, defaults to 1.0",
			header: "gzip;q=invalid, deflate",
			expected: map[string]float64{
				"gzip":    1.0,
				"deflate": 1.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAcceptEncoding(tt.header)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestChooseEncoding(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "Prefer gzip when both gzip and deflate are available",
			header:   "gzip, deflate",
			expected: "gzip",
		},
		{
			name:     "Choose deflate when gzip is not available",
			header:   "deflate;q=1.0",
			expected: "deflate",
		},
		{
			name:     "Return empty when no acceptable encodings are provided",
			header:   "br",
			expected: "",
		},
		{
			name:     "Return empty when quality is 0",
			header:   "gzip;q=0, deflate;q=0",
			expected: "",
		},
		{
			name:     "Ignore encoding with quality 0 and choose available",
			header:   "gzip;q=0, deflate;q=1.0",
			expected: "deflate",
		},
		{
			name:     "Handle empty Accept-Encoding header",
			header:   "",
			expected: "",
		},
		{
			name:     "Handle invalid quality value, default to gzip",
			header:   "gzip;q=invalid, deflate;q=0.8",
			expected: "gzip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := chooseEncoding(tt.header)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to decompress gzip data
func gzipDecompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return decompressed, nil
}

func TestGzipCompress(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "Non-empty input",
			input:   []byte("Hello, world!"),
			wantErr: false,
		},
		{
			name:    "Empty input",
			input:   []byte(""),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, encoding, err := gzipCompress(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("gzipCompress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Decompress the output for validation
			decompressed, err := gzipDecompress(got)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, decompressed)
			assert.Equal(t, "gzip", encoding)
		})
	}
}

// Helper function to decompress deflate data
func flateDecompress(data []byte) ([]byte, error) {
	reader := flate.NewReader(bytes.NewReader(data))
	defer reader.Close()
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return decompressed, nil
}

func TestDeflateCompress(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "Non-empty input",
			input:   []byte("Hello, world!"),
			wantErr: false,
		},
		{
			name:    "Empty input",
			input:   []byte(""),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, encoding, err := deflateCompress(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("deflateCompress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error, verify the output by decompressing
			if !tt.wantErr {
				decompressed, err := flateDecompress(got)
				assert.NoError(t, err)
				assert.Equal(t, tt.input, decompressed)
				assert.Equal(t, "deflate", encoding)
			}
		})
	}
}

func TestEncodeBody(t *testing.T) {
	tests := []struct {
		name                 string
		body                 string
		acceptEncodingHeader string
		wantErr              bool
		expectedOutput       []byte // nil when we expect an error
		expectedEncoding     string // empty when we expect an error
	}{
		{
			name:                 "Encode with gzip",
			body:                 "Hello, world!",
			acceptEncodingHeader: "gzip",
			wantErr:              false,
			expectedOutput:       []byte("Hello, world!"),
			expectedEncoding:     "gzip",
		},
		{
			name:                 "Encode with deflate",
			body:                 "Hello, world!",
			acceptEncodingHeader: "deflate",
			wantErr:              false,
			expectedOutput:       []byte("Hello, world!"),
			expectedEncoding:     "deflate",
		},
		{
			name:                 "No encoding",
			body:                 "Hello, world!",
			acceptEncodingHeader: "",
			wantErr:              false,
			expectedOutput:       []byte("Hello, world!"),
			expectedEncoding:     "",
		},
		{
			name:                 "identity encoding",
			body:                 "Hello, world!",
			acceptEncodingHeader: "identity",
			wantErr:              false,
			expectedOutput:       []byte("Hello, world!"),
			expectedEncoding:     "",
		},
		{
			name:                 "Unsupported encoding",
			body:                 "Hello, world!",
			acceptEncodingHeader: "br",
			wantErr:              false,
			expectedOutput:       []byte("Hello, world!"),
			expectedEncoding:     "",
		},
		{
			name:                 "Empty body with gzip",
			body:                 "",
			acceptEncodingHeader: "gzip",
			wantErr:              false,
			expectedOutput:       []byte(""),
			expectedEncoding:     "gzip",
		},
		{
			name:                 "Empty body with no encoding",
			body:                 "",
			acceptEncodingHeader: "",
			wantErr:              false,
			expectedOutput:       []byte(""),
			expectedEncoding:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, encoding, err := EncodeBody(&tt.body, tt.acceptEncodingHeader)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeBody() test = %s output = %s encoding = %s error = %v, wantErr %v", tt.name, output, encoding, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.expectedEncoding, encoding)

				if tt.acceptEncodingHeader == "gzip" || tt.acceptEncodingHeader == "deflate" {
					// Decompress and compare
					var decompressed []byte
					var decompressErr error
					if tt.acceptEncodingHeader == "gzip" {
						decompressed, decompressErr = gzipDecompress(output)
					} else if tt.acceptEncodingHeader == "deflate" {
						decompressed, decompressErr = flateDecompress(output)
					}
					assert.NoError(t, decompressErr)
					assert.Equal(t, tt.expectedOutput, decompressed)

				} else {
					// Directly compare the output
					assert.Equal(t, tt.expectedOutput, output)
				}
			}
		})
	}
}
