package config

// Config objects configure the proxy proxy.
type Config struct {
	httpBehavior
	terminalLogger
	trafficLogger
}

func (cfg *Config) SetLoggerLevel() {
	cfg.terminalLogger.setLoggerLevel()
}

func (cfg *Config) GetDebugLevel() int {
	return cfg.terminalLogger.getDebugLevel()
}

func NewDefaultConfig() *Config {
	return &Config{
		httpBehavior: httpBehavior{
			Listen:                "127.0.0.1:8080",
			CertDir:               "",
			InsecureSkipVerifyTLS: false,
			NoHttpUpgrader:        false,
		},
		terminalLogger: terminalLogger{
			Verbose: false,
			Debug:   false,
		},
		trafficLogger: trafficLogger{
			OutputDir:           "",
			WriteJsonFormatLogs: true,
		},
	}
}
