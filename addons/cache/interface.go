package cache

import (
	"github.com/robbyt/llm_proxy/schema"
)

type DB interface {
	Close() error
	Get(request *schema.TrafficObject) (response *schema.TrafficObject, err error)
	Put(request *schema.TrafficObject, response *schema.TrafficObject) error
}
