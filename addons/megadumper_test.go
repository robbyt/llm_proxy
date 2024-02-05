package addons

import (
	"net/url"
	"testing"

	"github.com/kardianos/mitmproxy/proxy"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	md "github.com/robbyt/llm_proxy/addons/megadumper"
)

type mockWriter struct {
	bucket []byte
}

func (w *mockWriter) Write(p []byte) (n int, err error) {
	w.bucket = append(w.bucket, p...)
	return -1, nil
}

func TestMegaDumper(t *testing.T) {
	dumper := &MegaDumper{
		singleLogFileTarget: nil,
		logFilename:         "",
		logTarget:           "",
		logLevel:            md.LogLevel(0),
		logFormat:           md.LogFormat(0),
	}

	assert.NotNil(t, dumper)
	assert.Implements(t, (*proxy.Addon)(nil), dumper)
}

func TestMegaDumper_SetLogLevel(t *testing.T) {
	dumper := &MegaDumper{}
	logLevel := md.LogLevel(1)
	dumper.logLevel = logLevel

	assert.Equal(t, logLevel, dumper.logLevel)
}

func TestMegaDumper_SetLogFormat(t *testing.T) {
	dumper := &MegaDumper{}
	logFormat := md.LogFormat(1)
	dumper.logFormat = logFormat

	assert.Equal(t, logFormat, dumper.logFormat)
}

func TestMegaDumper_SetLogTarget(t *testing.T) {
	dumper := &MegaDumper{}
	logTarget := "target.log"
	dumper.logTarget = logTarget

	assert.Equal(t, logTarget, dumper.logTarget)
}

func TestMegaDumper_SingleLogFileTarget(t *testing.T) {
	dumper := &MegaDumper{
		logLevel:  md.WRITE_REQ_HEADERS_AND_REQ_BODY_ALSO_RESP_HEADERS_AND_RESP_BODY,
		logFormat: md.LogFormat_PLAINTEXT,
	}
	logFileTarget := &mockWriter{}
	dumper.singleLogFileTarget = logFileTarget
	assert.Equal(t, logFileTarget, dumper.singleLogFileTarget)

	// create a fake flow, passed to the writer for full in-memory testing
	flow := &proxy.Flow{
		Id: uuid.FromStringOrNil("123"),
		Request: &proxy.Request{
			URL: &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
			},
		},
		Response: &proxy.Response{
			StatusCode: 200,
			Body:       []byte("Hello, World!"),
		},
	}
	dumper.Write(flow)

	// TODO check the in-memory log file

}

func TestLogExtension(t *testing.T) {
	tests := []struct {
		name     string
		format   md.LogFormat
		expected string
	}{
		{
			name:     "Plain Text",
			format:   md.LogFormat_PLAINTEXT,
			expected: "log",
		},
		{
			name:     "JSON",
			format:   md.LogFormat_JSON,
			expected: "json",
		},
		{
			name:     "Unknown",
			format:   md.LogFormat(99), // invalid enum
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mDump := &MegaDumper{
				logFormat: tt.format,
			}

			result := mDump.logExtension()
			assert.Equal(t, tt.expected, result)
		})
	}
}
