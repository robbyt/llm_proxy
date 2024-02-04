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
	dirLoggerCmd.Flags().StringVarP(
		&cfg.Listen, "listen", "l", cfg.Listen,
		"Address to listen on",
	)
	dirLoggerCmd.Flags().StringVarP(
		&cfg.CertDir, "ca_dir", "c", cfg.CertDir,
		"Path to the local trusted certificate, for TLS MITM",
	)
	dirLoggerCmd.Flags().BoolVarP(
		&cfg.InsecureSkipVerifyTLS, "skip-upstream-tls-verify", "K", cfg.InsecureSkipVerifyTLS,
		"Skip upstream TLS cert verification",
	)
	dirLoggerCmd.Flags().BoolVarP(
		&cfg.NoHttpUpgrader, "no-http-upgrader", "", cfg.NoHttpUpgrader,
		"Disable the http->https upgrader. If set, the proxy will not upgrade http requests to https.",
	)

	cfg.OutputDir = "/tmp/llm_proxy" // set a default output directory for this command
	dirLoggerCmd.Flags().StringVarP(
		&cfg.OutputDir, "output", "o", cfg.OutputDir,
		"Directory to write logs",
	)
}
