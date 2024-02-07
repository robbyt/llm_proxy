/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	log "github.com/sirupsen/logrus"
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
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(simpleCmd)
}
