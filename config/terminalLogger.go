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
func (tlo *terminalLogger) setLoggerLevel() {
	if tlo.Debug {
		log.SetLevel(log.DebugLevel)
		if tlo.Trace {
			log.SetReportCaller(true)
		}
	} else if tlo.Verbose {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
	log.Info("Logger level set to: ", log.GetLevel())
	tlo.logLevelHasBeenSet = true
}

// getDebugLevel returns 1 if the log level is debug, 0 otherwise, for use in the proxy package
func (tlo *terminalLogger) getDebugLevel() int {
	if !tlo.logLevelHasBeenSet {
		tlo.setLoggerLevel()
	}

	if log.GetLevel() >= log.DebugLevel {
		return 1
	}
	return 0
}
