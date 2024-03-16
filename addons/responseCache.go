package addons

import (
	"fmt"

	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"

	"github.com/robbyt/llm_proxy/addons/cache"
	"github.com/robbyt/llm_proxy/addons/megadumper/formatters"
)

type ResponseCacheAddon struct {
	px.BaseAddon
	formatter formatters.MegaDumpFormatter
	cache     cache.DB
	cacheHit  bool
}

func (mca *ResponseCacheAddon) Request(f *px.Flow) {
	mca.cacheHit = false
	// Only cache these request methods (and empty string for GET)
	cacheOnlyMethods := map[string]struct{}{
		"GET":     {},
		"":        {},
		"HEAD":    {},
		"OPTIONS": {},
		"POST":    {},
	}
	if _, ok := cacheOnlyMethods[f.Request.Method]; !ok {
		log.Debugf("skipping cache for: %s", f.Request.URL)
		return
	}

	cacheLookup, err := mca.cache.Lookup(*f.Request)
	if err != nil {
		log.Errorf("error accessing cache, bypassing: %s", err)
		return
	}
	if cacheLookup != nil {
		log.Debugf("cache hit for: %s", f.Request.URL)
		mca.cacheHit = true
		f.Response = cacheLookup
		// TODO add a response header to indicate this is a cache hit
	}
}

func (c *ResponseCacheAddon) Response(f *px.Flow) {
	go func() {
		<-f.Done()
		if err := c.cache.Store(*f.Request, f.Response); err != nil {
			log.Errorf("error storing response in cache: %s", err)
		}
	}()
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
		cacheDB, err = cache.NewBadgerDB(cacheDir)
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
