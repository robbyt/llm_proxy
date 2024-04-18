package cmd

import (
	"github.com/spf13/cobra"

	"github.com/robbyt/llm_proxy/config"
	"github.com/robbyt/llm_proxy/proxy"
)

// simpleCmd represents a simple proxy server without logging
var simpleCmd = &cobra.Command{
	Use:   "simple",
	Short: "Run a 'simple' proxy server, traffic will not be stored or cached.",
	Long:  "Useful as a simple proxy to test connectivity, or as a base for a more complex proxy.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.AppMode = config.SimpleMode
		return proxy.Run(cfg)
	},
}

func init() {
	rootCmd.AddCommand(simpleCmd)
	simpleCmd.SuggestFor = simple_suggestions
}
