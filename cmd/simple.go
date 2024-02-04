/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/robbyt/llm_proxy/proxy"
)

// simpleCmd represents a simple proxy server without logging
var simpleCmd = &cobra.Command{
	Use:   "simple",
	Short: "Run a 'simple' proxy server, request logs will not be written to disk",
	Long: `Useful as a simple proxy, or as a base for a more complex proxy.
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := proxy.Run(cfg)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(simpleCmd)
	simpleCmd.Flags().StringVarP(&cfg.Listen, "listen", "l", cfg.Listen, "Address to listen on")
	simpleCmd.Flags().StringVarP(&cfg.CertDir, "ca_dir", "c", cfg.CertDir, "Path to the local trusted certificate, for TLS MITM")
	simpleCmd.Flags().BoolVarP(&cfg.InsecureSkipVerifyTLS, "no-upstream-tls-verify", "K", cfg.InsecureSkipVerifyTLS, "Skip upstream TLS cert verification")
	simpleCmd.Flags().BoolVarP(&cfg.NoHttpUpgrader, "no-http-upgrader", "", cfg.NoHttpUpgrader, "Disable the http->https upgrader. If set, the proxy will not upgrade http requests to https.")
}
