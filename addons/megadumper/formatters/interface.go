package formatters

import md "github.com/robbyt/llm_proxy/addons/megadumper"

// MegaDumpFormatter abstracts the different types of log storage formats
type MegaDumpFormatter interface {
	Read(container *md.LogDumpContainer) ([]byte, error)
}
