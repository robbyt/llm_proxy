/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var verbose bool
var debug bool

func setLoggerLevel() {
	if debug {
		log.SetLevel(log.DebugLevel)
		// enable this for full code tracing output
		// log.SetReportCaller(true)
	} else if verbose {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
	log.Info("Logger level set to: ", log.GetLevel())
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "llm_proxy",
	Short: "Proxy your LLM traffic for logging, security evaluation, and fine-tuning.",
	Long: `llm_proxy is a HTTP MITM proxy that logs all requests and responses.

This is useful for:
  * security (this daemon can be run in a DMZ to bridge isolated applications and external LLM providers)
  * debugging (track all LLM traffic, and review it later if your app is producing unexpected results)
  * fine tuning (save all requests and responses to use as fine tuning data, which can improve LLM performance and accuracy)
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setLoggerLevel()
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
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print runtime activity to stdout")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Print debug information to stderr")
}
