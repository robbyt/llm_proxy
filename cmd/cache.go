/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/robbyt/llm_proxy/config"
	"github.com/robbyt/llm_proxy/proxy"
)

// cacheCmd represents the mock command
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Creates a caching server, storing and responding with previously generated responses",
	Long: `This command creates a proxy server that sends responses to the upstream server only
when there isn't a copy available in the cache. The cache command requires a local directory to store
and retrieve the responses. This mode is useful for development and for CI, because it will reduce the
number of requests to the upstream server. The cache server will respond with the same status code,
headers, and body as the previous response. The cache server will not store responses with a status
code of 500 or higher.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg.AppMode = config.CacheMode
		err := proxy.Run(cfg)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.Flags().StringVarP(
		&cfg.Cache.Dir, "cache", "o", cfg.Cache.Dir,
		"Directory to store the cache files",
	)
	/*
		cacheCmd.Flags().Int64VarP(
			&cfg.Cache.TTL, "ttl", "", cfg.Cache.TTL,
			"Time to live for cache files in seconds (0 means cache forever)",
		)
	*/
	cacheCmd.Flags().StringSliceVarP(
		&cfg.FilterReqHeaders, "filter-req-headers", "", cfg.FilterReqHeaders,
		"Request headers that match these strings will not be logged (but will still be proxied)",
	)
	cacheCmd.Flags().StringSliceVarP(
		&cfg.FilterRespHeaders, "filter-resp-headers", "", []string{},
		"Response headers that match these strings will not be logged (but will still be proxied)",
	)
}
