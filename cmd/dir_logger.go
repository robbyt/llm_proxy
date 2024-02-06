/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/robbyt/llm_proxy/proxy"
)

// dirLoggerCmd represents the demo command
var dirLoggerCmd = &cobra.Command{
	Use:   "dir_logger",
	Short: "Proxy requests and write logs to a directory on disk",
	Long: `Run a proxy server, and write all requests and responses to disk.
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := proxy.Run(cfg)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(dirLoggerCmd)

	// setting the default value here instead of in the config struct, because setting this
	// to _something_ reconfigures the output, to write to a directory instead of a single file.
	cfg.OutputDir = "/tmp/llm_proxy"

	dirLoggerCmd.Flags().StringVarP(
		&cfg.OutputDir, "output", "o", cfg.OutputDir,
		"Directory to write logs",
	)
}
