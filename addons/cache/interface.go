package cache

import (
	px "github.com/kardianos/mitmproxy/proxy"
)

type DB interface {
	Close() error
	// GetStr(url string, key string) ([]byte, error)
	// GetBytes(url string, key []byte) ([]byte, error)
	// SetStr(url string, key string, value []byte) error
	// SetBytes(url string, key []byte, value []byte) error
	Get(req px.Request) (*px.Response, error)
	Put(req px.Request, resp *px.Response) error
}
