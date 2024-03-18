package cache

import (
	"net/http"
	"net/url"
	"testing"

	px "github.com/kardianos/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBadgerMetaDB(t *testing.T) {
	tempDir := t.TempDir()

	metaDB, err := NewBadgerMetaDB(tempDir)
	require.NoError(t, err)
	require.NotNil(t, metaDB)

	// Check the dbCache property
	assert.NotNil(t, metaDB.dbCache)

	// Clean up
	err = metaDB.Close()
	assert.NoError(t, err)
}

func TestBadgerMetaDB_Store(t *testing.T) {
	tempDir := t.TempDir()

	metaDB, _ := NewBadgerMetaDB(tempDir)
	defer metaDB.Close()

	req := px.Request{
		URL:  &url.URL{Scheme: "http", Host: "localhost", Path: "/test"},
		Body: []byte("test body"),
	}
	resp := px.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body:       []byte("test response"),
	}

	// Empty DB, test retrieving a non-existent request
	storedValue, err := metaDB.Get(req)
	require.NoError(t, err)
	require.Nil(t, storedValue)

	// Test storing a request and response
	err = metaDB.Put(req, &resp)
	require.NoError(t, err)

	// Test retrieving the stored request
	storedValue, err = metaDB.Get(req)
	require.NoError(t, err)
	require.NotNil(t, storedValue)

	// The stored value should match the original request body
	assert.Equal(t, resp.StatusCode, storedValue.StatusCode)
	assert.Equal(t, resp.Header, storedValue.Header)
	assert.Equal(t, resp.Body, storedValue.Body)

	// Clean up
	err = metaDB.Close()
	assert.NoError(t, err)
}

func TestBadgerMetaDB_StoreEmptyBody(t *testing.T) {
	tempDir := t.TempDir()

	metaDB, _ := NewBadgerMetaDB(tempDir)
	defer metaDB.Close()

	req := px.Request{
		URL:  &url.URL{Scheme: "http", Host: "localhost", Path: "/test"},
		Body: []byte(""),
	}

	resp := px.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body:       []byte("test response"),
	}

	// Empty DB, test retrieving a non-existent request
	storedValue, err := metaDB.Get(req)
	require.NoError(t, err)
	require.Nil(t, storedValue)

	// Test storing a request with an empty body
	err = metaDB.Put(req, &resp)
	require.NoError(t, err)

	// The stored value should be a single space, as per the implementation
	storedValue, err = metaDB.Get(req)
	require.NoError(t, err)
	require.NotNil(t, storedValue)

	// The stored value should match the original request body
	assert.Equal(t, resp.StatusCode, storedValue.StatusCode)
	assert.Equal(t, resp.Header, storedValue.Header)
	assert.Equal(t, resp.Body, storedValue.Body)

	// Clean up
	err = metaDB.Close()
	assert.NoError(t, err)
}
