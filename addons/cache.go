package addons

import (
	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"

	"github.com/robbyt/llm_proxy/addons/cache"
	"github.com/robbyt/llm_proxy/addons/cache/storage"
	"github.com/robbyt/llm_proxy/addons/megadumper/formatters"
)

type MockCacheAddon struct {
	px.BaseAddon
	logTarget string
	formatter formatters.MegaDumpFormatter
	indexFile *cache.CacheIndexFile
	cache     storage.CacheDB
}

func (mca *MockCacheAddon) Request(f *px.Flow) {
	// Only cache these request methods
	cacheOnlyMethods := []string{"GET", "HEAD", "OPTIONS", "POST"}
	for _, method := range cacheOnlyMethods {
		if f.Request.Method != method {
			log.Debugf("skipping cache lookup for method: %s", f.Request.Method)
			return
		}
	}

	cacheLookup, err := mca.cache.Lookup(*f.Request)
	if err != nil {
		log.Errorf("error accessing cache, bypassing: %s", err)
		return
	}
	if cacheLookup != nil {
		log.Debugf("cache hit for: %s", f.Request.URL)
		f.Response = cacheLookup
	}
}

func NewCacheAddon(
	cacheDir string, // output & cache storage directory
	filterReqHeaders, filterRespHeaders []string, // which headers to filter out
) (*MockCacheAddon, error) {

	// load or create the cache index file, which points to the cache storage
	iFile, err := cache.NewCacheIndex(cacheDir)
	if err != nil {
		return nil, err
	}

	cacheDB, err := iFile.GetStorageEngine()
	if err != nil {
		return nil, err
	}

	return &MockCacheAddon{
		logTarget: cacheDir,
		formatter: &formatters.JSON{},
		indexFile: iFile,
		cache:     cacheDB,
	}, nil
}
