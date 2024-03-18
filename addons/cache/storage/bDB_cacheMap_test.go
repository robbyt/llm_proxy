package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBadgerDB_CacheMap(t *testing.T) {
	cacheMap := NewBadgerDB_CacheMap()
	assert.NotNil(t, cacheMap)
	assert.Equal(t, 0, cacheMap.Len())
}

func TestBadgerDB_CacheMap_PutAndGet(t *testing.T) {
	cacheMap := NewBadgerDB_CacheMap()
	tempDir := t.TempDir()

	badgerDB, _ := NewBadgerDB("test", tempDir+"/test.db")
	defer badgerDB.Close()

	cacheMap.Put("test", badgerDB)
	assert.Equal(t, 1, cacheMap.Len())

	retrievedDB, found := cacheMap.Get("test")
	assert.True(t, found)
	assert.Equal(t, badgerDB, retrievedDB)

	// access the entire map using All()
	all := cacheMap.All()
	assert.Equal(t, 1, len(all))
	assert.Equal(t, badgerDB, all["test"])
}

func TestBadgerDB_CacheMap_DeleteAndClear(t *testing.T) {
	cacheMap := NewBadgerDB_CacheMap()
	tempDir := t.TempDir()

	badgerDB, _ := NewBadgerDB("test", tempDir+"/test.db")
	defer badgerDB.Close()

	cacheMap.Put("test", badgerDB)
	assert.Equal(t, 1, cacheMap.Len())

	cacheMap.Delete("test")
	assert.Equal(t, 0, cacheMap.Len())

	cacheMap.Put("test", badgerDB)
	assert.Equal(t, 1, cacheMap.Len())

	cacheMap.Clear()
	assert.Equal(t, 0, cacheMap.Len())
}
