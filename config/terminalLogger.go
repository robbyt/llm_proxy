package config

import log "github.com/sirupsen/logrus"

// terminalLogger controls the logging output to the terminal while the proxy is running
type terminalLogger struct {
	Verbose            bool // if true, print runtime activity to stderr
	Debug              bool // if true, print debug information to stderr
	Trace              bool // if true, print detailed report caller tracing to stderr, for debugging
	logLevelHasBeenSet bool // internal flag to track if the log level has been set
}

// setLoggerLevel sets the logrus level based on verbose/debug values in the config object
func (ocfg *terminalLogger) setLoggerLevel() {
	if ocfg.Debug {
		log.SetLevel(log.DebugLevel)
		if ocfg.Trace {
			log.SetReportCaller(true)
		}
	} else if ocfg.Verbose {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
	log.Info("Logger level set to: ", log.GetLevel())
	ocfg.logLevelHasBeenSet = true
}

// getDebugLevel returns 1 if the log level is debug, 0 otherwise, for use in the proxy package
func (ocfg *terminalLogger) getDebugLevel() int {
	if !ocfg.logLevelHasBeenSet {
		ocfg.setLoggerLevel()
	}

	if log.GetLevel() >= log.DebugLevel {
		return 1
	}
	return 0
}
