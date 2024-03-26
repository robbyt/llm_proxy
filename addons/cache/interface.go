package cache

import (
	"github.com/robbyt/llm_proxy/schema"
)

type DB interface {
	Close() error
	Len(identifier string) (int, error)
	Get(identifier string, body []byte) (response *schema.TrafficObject, err error)
	Put(request *schema.TrafficObject, response *schema.TrafficObject) error
}
