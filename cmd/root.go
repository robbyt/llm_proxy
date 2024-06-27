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
  * Security: A multi-homed DMZ provides isolation between apps and external LLM APIs.
  * Debugging: Tag and observe all LLM API traffic.
  * Fine-tuning: Use the stored logs to fine-tune your LLM models.
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg.SetLoggerLevel()
	},
	SilenceUsage: true,
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
	rootCmd.PersistentFlags().BoolVar(
		&cfg.Trace, "trace", cfg.Trace, "Print detailed trace debugging information to stderr, requires --debug to also be set")
	rootCmd.PersistentFlags().MarkHidden("trace")

	rootCmd.Flags().StringVarP(
		&cfg.Listen, "listen", "l", cfg.Listen,
		"Address to listen on",
	)
	rootCmd.Flags().StringVarP(
		&cfg.CertDir, "ca_dir", "c", cfg.CertDir,
		"Path to the local trusted certificate, for TLS MITM",
	)
	rootCmd.Flags().BoolVarP(
		&cfg.InsecureSkipVerifyTLS, "skip-upstream-tls-verify", "K", cfg.InsecureSkipVerifyTLS,
		"Skip upstream TLS cert verification",
	)
	rootCmd.Flags().BoolVarP(
		&cfg.NoHttpUpgrader, "no-http-upgrader", "", cfg.NoHttpUpgrader,
		"Disable the automatic http->https request upgrader",
	)
}
