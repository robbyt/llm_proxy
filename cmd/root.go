/*
MIT License

Copyright (c) 2024 Robert Terhaar <robbyt@robbyt.net> All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package cmd

import (
	"os"

	"github.com/robbyt/llm_proxy/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func setLoggerLevel(cfg *config.Config) {
	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
		// enable this for full code tracing output
		// log.SetReportCaller(true)
	} else if cfg.Verbose {
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
		setLoggerLevel(cfg)
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
	rootCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Print runtime activity to stdout")
	rootCmd.PersistentFlags().BoolVarP(&cfg.Debug, "debug", "d", false, "Print debug information to stderr")
}
