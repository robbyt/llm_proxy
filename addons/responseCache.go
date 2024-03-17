package addons

import (
	"fmt"
	"sync/atomic"

	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"

	"github.com/robbyt/llm_proxy/addons/cache"
	"github.com/robbyt/llm_proxy/addons/megadumper/formatters"
)

var cacheOnlyMethods = map[string]struct{}{
	"GET":     {},
	"":        {},
	"HEAD":    {},
	"OPTIONS": {},
	"POST":    {},
}

type ResponseCacheAddon struct {
	px.BaseAddon
	formatter formatters.MegaDumpFormatter
	cache     cache.DB
	closed    atomic.Bool
}

func (mca *ResponseCacheAddon) Request(f *px.Flow) {
	// Only cache these request methods (and empty string for GET)
	if _, ok := cacheOnlyMethods[f.Request.Method]; !ok {
		log.Debugf(
			"skipping cache lookup for unsupported method: %s %s", f.Request.Method, f.Request.URL)
		return
	}

	cacheLookup, err := mca.cache.Lookup(*f.Request)
	if err != nil {
		log.Errorf("error accessing cache, bypassing: %s", err)
		return
	}
	if cacheLookup != nil {
		log.Debugf("cache hit for: %s", f.Request.URL)
		f.Response = cacheLookup
		// TODO add a response header to indicate this is a cache hit
		return
	}
	log.Debugf("cache miss for: %s", f.Request.URL)
}

func (c *ResponseCacheAddon) Response(f *px.Flow) {
	go func() {
		<-f.Done()
		if err := c.cache.Store(*f.Request, f.Response); err != nil {
			log.Errorf("error storing response in cache: %s", err)
		}
	}()
}

func (d *ResponseCacheAddon) String() string {
	return "ResponseCacheAddon"
}

func (d *ResponseCacheAddon) Close() error {
	if !d.closed.Swap(true) {
		log.Debug("Waiting for ResponseCacheAddon shutdown...")
		return d.cache.Close()
	}

	return nil
}

func NewCacheAddon(
	storageEngineName string, // name of the storage engine to use
	cacheDir string, // output & cache storage directory
	filterReqHeaders, filterRespHeaders []string, // which headers to filter out
) (*ResponseCacheAddon, error) {
	var cacheDB cache.DB
	var err error

	switch storageEngineName {
	case "badger":
		cacheDB, err = cache.NewBadgerMetaDB(cacheDir)
	default:
		return nil, fmt.Errorf("unknown storage engine: %s", storageEngineName)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating cache: %s", err)
	}

	return &ResponseCacheAddon{
		formatter: &formatters.JSON{},
		cache:     cacheDB,
	}, nil
}
