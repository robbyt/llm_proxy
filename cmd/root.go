package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "llm_proxy",
	Short: "Proxy your LLM traffic for logging, security evaluation, and fine-tuning.",
	Long: `llm_proxy is an HTTP MITM (Man-In-The-Middle) proxy designed to log all requests and responses.

This is useful for:
  * Security: The proxy daemon can operate in a DMZ to facilitate communication between isolated applications and external LLM API providers.
  * Debugging: It allows tracking all LLM API traffic, to enable later review if an application yields unexpected results.
  * Fine-tuning: By saving all requests and responses, this proxy allows the collection of fine-tuning data, which can be used to enhance LLM performance and accuracy.
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg.SetLoggerLevel()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true // don't show the default completion command in help
	rootCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", cfg.Verbose, "Print runtime activity to stderr")
	rootCmd.PersistentFlags().BoolVarP(&cfg.Debug, "debug", "d", cfg.Debug, "Print debug information to stderr")
	rootCmd.PersistentFlags().BoolVarP(&cfg.Trace, "trace", "", cfg.Trace, "Print detailed trace debugging information to stderr, requires --debug to also be set")

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
}
