package config

type cacheBehavior struct {
	Dir string // Directory to store the cache files
	TTL int64  // Time to live for cache files in seconds (0 means cache forever)
}
