package addons

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
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

	cacheLookup, err := mca.cache.Get(*f.Request)
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

var responseCodesToCache = []int{http.StatusOK, http.StatusAccepted, http.StatusCreated, http.StatusNoContent}

func (c *ResponseCacheAddon) Response(f *px.Flow) {
	go func() {
		<-f.Done()
		if f.Response == nil {
			log.Debugf("skipping cache storage for nil response: %s", f.Request.URL)
			return
		}

		shouldCache := false
		for _, respCode := range responseCodesToCache {
			if f.Response.StatusCode == respCode {
				shouldCache = true
				break
			}
		}

		if !shouldCache {
			log.Debugf("skipping cache storage for non-200 response: %s", f.Request.URL)
			return
		}

		if err := c.cache.Put(*f.Request, f.Response); err != nil {
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

func cleanCacheDir(cacheDir string) (string, error) {
	if cacheDir == "" {
		cacheDir = "."
	}

	invalidChars := []string{"<", ">", ":", "\"", "\\", "|", "?", "*", "!", "+", "`", "'"}
	for _, char := range invalidChars {
		if strings.Contains(cacheDir, char) {
			return "", fmt.Errorf("filename contains invalid character: %s", char)
		}
	}

	cacheDir, err := filepath.Abs(cacheDir)
	if err != nil {
		return "", err
	}

	return cacheDir, nil
}

func NewCacheAddon(
	storageEngineName string, // name of the storage engine to use
	cacheDir string, // output & cache storage directory
	filterReqHeaders, filterRespHeaders []string, // which headers to filter out
) (*ResponseCacheAddon, error) {
	var cacheDB cache.DB
	var err error
	cacheDir, err = cleanCacheDir(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("error cleaning cache path: %s", err)
	}

	switch storageEngineName {
	case "badger":
		// cacheDB, err = cache.NewBadgerMetaDB(cacheDir)
		panic("badger storage engine is disabled")
	case "bolt":
		cacheDB, err = cache.NewBoltMetaDB(cacheDir)
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
