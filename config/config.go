package config

// Config is the main config mega-struct
type Config struct {
	*httpBehavior
	*terminalLogger
	*trafficLogger
}

func (cfg *Config) getTerminalLogger() *terminalLogger {
	if cfg.terminalLogger == nil {
		cfg.terminalLogger = &terminalLogger{}
	}

	return cfg.terminalLogger
}
func (cfg *Config) SetLoggerLevel() {
	cfg.getTerminalLogger().setLoggerLevel()
}

func (cfg *Config) GetDebugLevel() int {
	return cfg.getTerminalLogger().getDebugLevel()
}

func NewDefaultConfig() *Config {
	return &Config{
		httpBehavior: &httpBehavior{
			Listen:                "127.0.0.1:8080",
			CertDir:               "",
			InsecureSkipVerifyTLS: false,
			NoHttpUpgrader:        false,
		},
		terminalLogger: &terminalLogger{},
		trafficLogger: &trafficLogger{
			OutputDir:           "",
			WriteJsonFormatLogs: true,
			FilterReqHeaders:    append([]string{}, defaultFilterHeaders...), // append empty to copy the slice from defaults
			FilterRespHeaders:   append([]string{}, defaultFilterHeaders...),
		},
	}
}
