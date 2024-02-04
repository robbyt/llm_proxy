package config

// Config objects configure the proxy proxy.
type Config struct {
	Listen    string // Local address the proxy should listen on
	OutputDir string // Directory to write logs
	CertDir   string // Dir to the certificate, for TLS MITM
	// SpoolDir              string // Directory to write files that have been "spooled" and pending uploading
	InsecureSkipVerifyTLS bool // if true, MITM will not verify the TLS certificate of the target server
	NoHttpUpgrader        bool // if true, the proxy will NOT upgrade http requests to https
	WriteJsonFormatLogs   bool // if true, write logs in JSON format
}

func GetDefaultConfig() *Config {
	return &Config{
		Listen:                "127.0.0.1:8080",
		OutputDir:             "",
		CertDir:               "",
		InsecureSkipVerifyTLS: false,
		NoHttpUpgrader:        false,
		WriteJsonFormatLogs:   true,
	}
}
