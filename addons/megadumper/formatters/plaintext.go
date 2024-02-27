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

	if pt.container.Request.Headers != nil {
		buf.WriteString(pt.container.Request.HeadersString())
	}

	if pt.container.Request.Body != "" {
		buf.WriteString(pt.container.Request.Body)
		buf.WriteString("\r\n")
	}

	if pt.container.Response.Headers != nil {
		buf.WriteString(pt.container.Response.HeadersString())
	}

	if pt.container.Response.Body != "" {
		buf.WriteString(pt.container.Response.Body)
		buf.WriteString("\r\n")
	}

	return buf.Bytes(), nil

}

// Read returns a flattened representation of all the fields in the LogDumpContainer
func (pt *PlainText) Read(container *schema.LogDumpContainer) ([]byte, error) {
	pt.container = container
	return pt.flatten()
}
