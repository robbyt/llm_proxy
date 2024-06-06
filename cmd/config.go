package cmd

import "github.com/proxati/llm_proxy/config"

// cfg is a reasonable default configuration, used by all commands
var cfg *config.Config = config.NewDefaultConfig()

// suggestions are here instead of their respective files bc it's easier to see them all in one place
var cache_suggestions = []string{
	"cache-proxy", "caching-proxy", "cash-proxy", "cash",
}

var dir_logger_suggestions = []string{
	"logger", "log", "dirlog", "dir-logger",
}

var simple_suggestions = []string{
	"proxy", "simple-proxy", "simpleproxy",
}
