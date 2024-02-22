package formatters

import "github.com/robbyt/llm_proxy/addons/megadumper/schema"

// MegaDumpFormatter abstracts the different types of log storage formats
type MegaDumpFormatter interface {
	Read(container *schema.LogDumpContainer) ([]byte, error)
}
