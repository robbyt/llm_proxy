package formatters

import (
	"encoding/json"
	"fmt"

	md "github.com/robbyt/llm_proxy/addons/megadumper"
)

type JSON struct {
	container *md.LogDumpContainer
}

// dumpToJSONBytes converts the requestLogDump struct to a byte array, omitting fields that are empty
func (f *JSON) dumpToJSONBytes() ([]byte, error) {
	if f.container == nil {
		return []byte("{}"), nil
	}

	j, err := json.MarshalIndent(f.container, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal requestLogDump to JSON: %w", err)
	}
	return j, nil
}

// Read returns the JSON representation of a LogDumpContainer (json formatted byte array)
func (f *JSON) Read(container *md.LogDumpContainer) ([]byte, error) {
	f.container = container
	return f.dumpToJSONBytes()
}
