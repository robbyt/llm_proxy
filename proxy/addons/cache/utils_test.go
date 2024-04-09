package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertIDtoFileName(t *testing.T) {
	tests := []struct {
		name      string
		dbFileDir string
		url       string
		want      string
	}{
		{
			name:      "Test with https URL",
			dbFileDir: "/tmp",
			url:       "https://example.com/test?param=value",
			want:      "/tmp/ZXhhbXBsZS5jb20vdGVzdD9wYXJhbT12YWx1ZQ==",
		},
		{
			name:      "Test with http URL",
			dbFileDir: "/tmp",
			url:       "http://example.com/test?param=value&param2=value2",
			want:      "/tmp/ZXhhbXBsZS5jb20vdGVzdD9wYXJhbT12YWx1ZSZwYXJhbTI9dmFsdWUy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertIDtoFileName(tt.dbFileDir, tt.url)
			assert.Equal(t, tt.want, got)
		})
	}
}
