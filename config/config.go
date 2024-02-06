package config

import (
	log "github.com/sirupsen/logrus"
)

// Config objects configure the proxy proxy.
type Config struct {
	Listen    string // Local address the proxy should listen on
	OutputDir string // Directory to write logs
	CertDir   string // Dir to the certificate, for TLS MITM
	// SpoolDir              string // Directory to write files that have been "spooled" and pending uploading
	InsecureSkipVerifyTLS bool // if true, MITM will not verify the TLS certificate of the target server
	NoHttpUpgrader        bool // if true, the proxy will NOT upgrade http requests to https
	WriteJsonFormatLogs   bool // if true, write logs in JSON format
	Verbose               bool // if true, print runtime activity to stdout
	Debug                 bool // if true, print debug information to stderr
	logLevelHasBeenSet    bool // internal flag to track if the log level has been set
}

// SetLoggerLevel sets the logrus level based on verbose/debug values in the config object
func (cfg *Config) SetLoggerLevel() {
	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
		// enable this for full code tracing output
		// log.SetReportCaller(true)
	} else if cfg.Verbose {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
	log.Info("Logger level set to: ", log.GetLevel())
	cfg.logLevelHasBeenSet = true
}

// GetDebugLevel returns 1 if the log level is debug, 0 otherwise, for use in the proxy package
func (cfg *Config) GetDebugLevel() int {
	if !cfg.logLevelHasBeenSet {
		cfg.SetLoggerLevel()
	}

	if log.GetLevel() >= log.DebugLevel {
		return 1
	}
	return 0
}

func GetDefaultConfig() *Config {
	return &Config{
		Listen:                "127.0.0.1:8080",
		OutputDir:             "",
		CertDir:               "",
		InsecureSkipVerifyTLS: false,
		NoHttpUpgrader:        false,
		WriteJsonFormatLogs:   true,
		Verbose:               false,
		Debug:                 false,
	}
}
