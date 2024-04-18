package cmd

import (
	"github.com/spf13/cobra"

	"github.com/robbyt/llm_proxy/config"
	"github.com/robbyt/llm_proxy/proxy"
)

// dirLoggerCmd represents the demo command
var dirLoggerCmd = &cobra.Command{
	Use:   "dir_logger",
	Short: "Proxy requests and write logs to a directory on disk",
	Long: `Run a proxy server, and write all requests and responses to a directory on disk.
Each request/response pair will be written to a file, identified by a unique ID. The file
will contain the request and response in JSON format.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.AppMode = config.DirLoggerMode
		return proxy.Run(cfg)
	},
}

func init() {
	rootCmd.AddCommand(dirLoggerCmd)
	dirLoggerCmd.SuggestFor = dir_logger_suggestions

	// setting the default value here instead of in the config struct factory, because setting
	// this to _something_ reconfigures the output, so it writes multi logs to a dir instead of
	// a single log to a file
	dirLoggerCmd.Flags().StringVarP(
		&cfg.OutputDir, "output", "o", "/tmp/llm_proxy",
		"Directory to write logs",
	)
	dirLoggerCmd.Flags().BoolVarP(
		&cfg.NoLogConnStats, "no-log-connection-stats", "", cfg.NoLogConnStats,
		"Don't log connection stats",
	)
	dirLoggerCmd.Flags().BoolVarP(
		&cfg.NoLogReqHeaders, "no-log-req-headers", "", cfg.NoLogReqHeaders,
		"Don't log request headers",
	)
	dirLoggerCmd.Flags().BoolVarP(
		&cfg.NoLogReqBody, "no-log-req-body", "", cfg.NoLogReqBody,
		"Don't log request body or details",
	)
	dirLoggerCmd.Flags().BoolVarP(
		&cfg.NoLogRespHeaders, "no-log-resp-headers", "", cfg.NoLogRespHeaders,
		"Don't log response headers",
	)
	dirLoggerCmd.Flags().BoolVarP(
		&cfg.NoLogRespBody, "no-log-resp-body", "", cfg.NoLogRespBody,
		"Don't log response body or details",
	)
	dirLoggerCmd.Flags().StringSliceVarP(
		&cfg.FilterReqHeaders, "filter-req-headers", "", cfg.FilterReqHeaders,
		"Request headers that match these strings will not be logged (but will still be proxied)",
	)
	dirLoggerCmd.Flags().StringSliceVarP(
		&cfg.FilterRespHeaders, "filter-resp-headers", "", cfg.FilterRespHeaders,
		"Response headers that match these strings will not be logged (but will still be proxied)",
	)
}
