package utils

import (
	"bytes"
	"net/http"
)

// HeaderString returns http.Header as a flat string
func HeaderString(headers http.Header) string {
	buf := new(bytes.Buffer)
	if err := headers.WriteSubset(buf, nil); err != nil {
		return ""
	}
	return buf.String()
}
