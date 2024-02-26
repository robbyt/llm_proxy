package formatters

import (
	"bytes"

	"github.com/robbyt/llm_proxy/schema"
)

type PlainText struct {
	container *schema.LogDumpContainer
}

func (pt *PlainText) flatten() ([]byte, error) {
	if pt.container == nil {
		return []byte(""), nil
	}

	buf := new(bytes.Buffer)

	if pt.container.RequestHeaders != nil {
		buf.WriteString(pt.container.RequestHeadersString())
	}

	if pt.container.RequestBody != "" {
		buf.WriteString(pt.container.RequestBody)
		buf.WriteString("\r\n")
	}

	if pt.container.ResponseHeaders != nil {
		buf.WriteString(pt.container.ResponseHeadersString())
	}

	if pt.container.ResponseBody != "" {
		buf.WriteString(pt.container.ResponseBody)
		buf.WriteString("\r\n")
	}

	return buf.Bytes(), nil

}

// Read returns a flattened representation of all the fields in the LogDumpContainer
func (pt *PlainText) Read(container *schema.LogDumpContainer) ([]byte, error) {
	pt.container = container
	return pt.flatten()
}
