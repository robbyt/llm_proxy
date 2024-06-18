package config

type AppMode int

const (
	SimpleMode AppMode = iota
	DirLoggerMode
	CacheMode
	APIAuditMode
)
