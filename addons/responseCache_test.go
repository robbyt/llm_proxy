package addons

import (
	"net/http"
	"net/url"
	"os"
	"testing"

	px "github.com/kardianos/mitmproxy/proxy"
	"github.com/robbyt/llm_proxy/schema"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanCachePath(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("empty cacheDir", func(t *testing.T) {
		cacheDir, err := cleanCacheDir("")
		require.Nil(t, err)
		assert.Equal(t, currentDir, cacheDir)
	})

	t.Run(". cacheDir", func(t *testing.T) {
		cacheDir, err := cleanCacheDir(".")
		require.Nil(t, err)
		assert.Equal(t, currentDir, cacheDir)
	})

	t.Run("non-empty cacheDir", func(t *testing.T) {
		cacheDir, err := cleanCacheDir("/tmp")
		require.Nil(t, err)
		assert.Equal(t, "/tmp", cacheDir)
	})

	t.Run("invalid cacheDir", func(t *testing.T) {
		cacheDir, err := cleanCacheDir("\\\\invalid\\path")
		require.NotNil(t, err)
		assert.Equal(t, "", cacheDir)
	})

	t.Run("relative cacheDir", func(t *testing.T) {
		cacheDir, err := cleanCacheDir("../../../../../../../../../../../tmp")
		assert.Nil(t, err)
		assert.Equal(t, "/tmp", cacheDir)
	})
}

func TestNewCacheAddonErr(t *testing.T) {
	filterReqHeaders := []string{"header1", "header2"}
	filterRespHeaders := []string{"header1", "header2"}

	t.Run("empty storage engine", func(t *testing.T) {
		storageEngineName := ""
		cacheDir := t.TempDir()
		cache, err := NewCacheAddon(storageEngineName, cacheDir, filterReqHeaders, filterRespHeaders)
		assert.Error(t, err, "Expected error for empty storage engine")
		assert.Nil(t, cache)
	})

	t.Run("unknown storage engine", func(t *testing.T) {
		storageEngineName := "unknown"
		cacheDir := t.TempDir()
		cache, err := NewCacheAddon(storageEngineName, cacheDir, filterReqHeaders, filterRespHeaders)
		assert.Error(t, err, "Expected error for unknown storage engine")
		assert.Nil(t, cache)
	})

	t.Run("bolt storage engine with invalid cacheDir", func(t *testing.T) {
		storageEngineName := "bolt"
		cacheDir := "\\\\invalid\\path"
		cache, err := NewCacheAddon(storageEngineName, cacheDir, filterReqHeaders, filterRespHeaders)
		assert.Error(t, err, "Expected error for invalid cacheDir")
		assert.Nil(t, cache)
	})

	t.Run("bolt storage engine with valid cacheDir", func(t *testing.T) {
		storageEngineName := "bolt"
		cacheDir := t.TempDir()
		cache, err := NewCacheAddon(storageEngineName, cacheDir, filterReqHeaders, filterRespHeaders)
		assert.NoError(t, err, "Expected no error for valid cacheDir")
		assert.NotNil(t, cache)
		assert.Equal(t, "ResponseCacheAddon", cache.String())
	})
}

func TestRequest(t *testing.T) {
	tmpDir := t.TempDir()
	filterReqHeaders := []string{"Header1"}
	filterRespHeaders := []string{"Header2"}
	respCacheAddon, err := NewCacheAddon(
		"bolt", tmpDir,
		filterReqHeaders, filterRespHeaders,
	)
	require.Nil(t, err)

	t.Run("cache miss", func(t *testing.T) {
		// first request made with an empty request body
		flow := &px.Flow{
			Request: &px.Request{
				Method: "POST",
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{
					"Host":    []string{"example.com"},
					"header1": []string{"value1"},
				},
				Body: []byte("req"),
			},
		}
		require.Empty(t, flow.Request.Header.Get("X-Cache"))

		// simulate the request hitting the addon
		respCacheAddon.Request(flow)
		require.Nil(t, flow.Response) // nil response means cache miss
		require.NotEmpty(t, flow.Request.Header.Get("X-Cache"))
		assert.Equal(t, "MISS", flow.Request.Header.Get("X-Cache"))
	})

	t.Run("cache hit", func(t *testing.T) {
		// simulate a response populating the cache after a miss
		flow := &px.Flow{
			Request: &px.Request{
				Method: "POST",
				URL:    &url.URL{Path: "/test"},
				Header: http.Header{
					"Host":    []string{"example.com"},
					"Header1": []string{"value1"},
					"Header2": []string{"value2"},
				},
				Body: []byte("req"),
			},
		}
		resp := &px.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"text/plain"},
				"Header1":      []string{"value1"},
				"Header2":      []string{"value2"},
			},
			Body: []byte("resp"),
		}
		identifier := flow.Request.URL.String()

		// unable to check length, because bucket doesn't yet exist in cache
		len, err := respCacheAddon.cache.Len(flow.Request.URL.String())
		require.Error(t, err)
		require.Zero(t, len)

		// create traffic objects for the request and response, check filtering
		tReq := schema.NewFromProxyRequest(flow.Request, filterReqHeaders)
		require.Empty(t, tReq.Header.Get("X-Cache"))
		require.Empty(t, tReq.Header.Get("header1"))    // deleted by filterReqHeaders
		require.NotEmpty(t, tReq.Header.Get("header2")) // not deleted by filterReqHeaders

		tResp := schema.NewFromProxyResponse(resp, filterRespHeaders)
		require.Empty(t, tResp.Header.Get("X-Cache"))
		require.NotEmpty(t, tResp.Header.Get("header1")) // not deleted by filterRespHeaders
		require.Empty(t, tResp.Header.Get("header2"))    // deleted by filterRespHeaders

		// store the response in cache using an internal method, to simulate the real response storage
		respCacheAddon.cache.Put(tReq, tResp)

		// check length again, should work now
		len, err = respCacheAddon.cache.Len(identifier)
		require.NoError(t, err)
		require.Equal(t, 1, len)

		// simulate a new request with the same URL, should be a hit now that it's in the cache
		require.Empty(t, resp.Header.Get("X-Cache"))
		respCacheAddon.Request(flow)
		require.NotNil(t, flow.Response)
		assert.Equal(t, resp.StatusCode, flow.Response.StatusCode)
		assert.Equal(t, resp.Body, flow.Response.Body)
		assert.Equal(t, "HIT", flow.Response.Header.Get("X-Cache"))

	})

	t.Cleanup(func() {
		respCacheAddon.Close()
	})
}
