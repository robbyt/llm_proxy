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

	t.Run("unknown storage engine", func(t *testing.T) {
		storageEngineName := "unknown"
		cacheDir := t.TempDir()
		_, err := NewCacheAddon(storageEngineName, cacheDir, filterReqHeaders, filterRespHeaders)
		assert.NotNil(t, err, "Expected error for unknown storage engine")
	})

	t.Run("bolt storage engine with invalid cacheDir", func(t *testing.T) {
		storageEngineName := "bolt"
		cacheDir := "\\\\invalid\\path"
		_, err := NewCacheAddon(storageEngineName, cacheDir, filterReqHeaders, filterRespHeaders)
		assert.NotNil(t, err, "Expected error for invalid cacheDir")
	})

	t.Run("bolt storage engine with valid cacheDir", func(t *testing.T) {
		storageEngineName := "bolt"
		cacheDir := t.TempDir()
		_, err := NewCacheAddon(storageEngineName, cacheDir, filterReqHeaders, filterRespHeaders)
		assert.Nil(t, err, "Expected no error for valid cacheDir")
	})
}

func TestRequest(t *testing.T) {
	/*
		reqEmptyBody := &px.Request{
			Method: "GET",
			URL:    &url.URL{Path: "/test"},
			Header: http.Header{"Host": []string{"example.com"}},
			Body:   []byte(""),
		}

		reqNormBodyURL1 := &px.Request{
			Method: "GET",
			URL:    &url.URL{Path: "/test"},
			Header: http.Header{"Host": []string{"example.com"}},
			Body:   []byte("hello"),
		}

		reqNormBodyURL2 := &px.Request{
			Method: "GET",
			URL:    &url.URL{Path: "/test2"},
			Header: http.Header{"Host": []string{"example.com"}},
			Body:   []byte("hello"),
		}

		respEmpty := &px.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Body:       []byte(""),
		}

		respNormal := &px.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Body:       []byte("hello"),
		}

		respErr := &px.Response{
			StatusCode: 500,
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Body:       []byte("error"),
		}
	*/

	storageEngine := "bolt"
	tmpDir := t.TempDir()
	filterReqHeaders := []string{"header1", "header2"}
	filterRespHeaders := []string{"header1", "header2"}
	respCacheAddon, err := NewCacheAddon(storageEngine, tmpDir, filterReqHeaders, filterRespHeaders)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("typical cached request for GET", func(t *testing.T) {
		req := &px.Request{
			Method: "GET",
			URL:    &url.URL{Path: "/test"},
			Header: http.Header{"Host": []string{"example.com"}},
			Body:   []byte(""),
		}
		tObjReq := schema.NewFromProxyRequest(req, filterReqHeaders)

		resp := &px.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Body:       []byte("hello"),
		}
		tObjResp := schema.NewFromProxyResponse(resp, filterRespHeaders)

		// create a new flow with only the request
		flow := &px.Flow{Request: req}

		// first request made with an empty request body
		respCacheAddon.Request(flow)
		require.Nil(t, flow.Response) // nil response means cache miss, as expected
		assert.Equal(t, "MISS", flow.Request.Header.Get("X-Cache"))

		// store a fake response in cache, to simulate a response populating the cache after a miss
		flow.Response = resp
		respCacheAddon.cache.Put(tObjReq, tObjResp)

		// not nil means the response was cached
		respCacheAddon.Request(flow)
		require.NotNil(t, flow.Response)
		assert.Equal(t, http.StatusOK, flow.Response.StatusCode)
		assert.Equal(t, "text/plain", flow.Response.Header.Get("Content-Type"))
		assert.Equal(t, "hello", string(flow.Response.Body))

	})
}
