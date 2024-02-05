package megadumper

import (
	"encoding/json"
	"fmt"
)

// LogDumpDiskContainer is a struct for holding the various types of data contained within a single request/response
type LogDumpDiskContainer struct {
	RequestHeaders  string `json:"request_headers,omitempty"`
	RequestBody     string `json:"request_body,omitempty"`
	ResponseHeaders string `json:"response_headers,omitempty"`
	ResponseBody    string `json:"response_body,omitempty"`
}

// DumpToJSONBytes converts the requestLogDump struct to a byte array, omitting fields that are empty
func (d *LogDumpDiskContainer) DumpToJSONBytes() ([]byte, error) {
	j, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal requestLogDump to JSON: %w", err)
	}
	return j, nil
}
