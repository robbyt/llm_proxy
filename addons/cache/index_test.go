package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIndexFile(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	tempFilePath := filepath.Join(tempDir, "tempFile.json")
	defer os.Remove(tempFilePath)

	// Write some text line to the file
	text := []byte(`{
		"schema_version": "v1",
		"cache_storage": {
			"name": "test",
			"path": "/path/to/storage",
			"storage_engine": "badger",
			"storage_version": "v1"
		}
	}`)
	err := os.WriteFile(tempFilePath, text, 0644)
	if err != nil {
		t.Fatalf("Failed to write to temporary file: %s", err)
	}

	// Use the NewIndexFile function to load the data from the file
	indexFile, err := NewCacheIndex(tempFilePath)

	// Assert that the data was loaded correctly
	assert.NotNil(t, indexFile)
	assert.Nil(t, err)
	assert.Equal(t, "v1", indexFile.SchemaVersion)
	assert.Equal(t, "test", indexFile.CacheStorage.Name)
	assert.Equal(t, "/path/to/storage", indexFile.CacheStorage.Path)
	assert.Equal(t, "badger", indexFile.CacheStorage.StorageEngine)
	assert.Equal(t, "v1", indexFile.CacheStorage.StorageVersion)

	// Test with a non-existing file
	indexFile, err = NewCacheIndex("/path/to/non/existing/file")
	assert.Equal(t, indexFile.FilePath, "/path/to/non/existing/file")
	assert.Nil(t, err)
}

func TestNewIndexFile_PermissionDenied(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	tempFilePath := filepath.Join(tempDir, "tempFile.json")
	defer os.Remove(tempFilePath)

	// Write some text line to the file
	text := []byte(`{
		"schema_version": "v1",
		"cache_storage": {
			"name": "test",
			"path": "/path/to/storage",
			"storage_engine": "badger",
			"storage_version": "v1"
		}
	}`)
	err := os.WriteFile(tempFilePath, text, 0000)
	if err != nil {
		t.Fatalf("Failed to write to temporary file: %s", err)
	}

	// Use the NewIndexFile function to load the data from the file
	indexFile, err := NewCacheIndex(tempFilePath)

	// Assert that the data was loaded correctly
	assert.Nil(t, indexFile)
	assert.NotNil(t, err)
}

func TestNewIndexFile_BadJson(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	tempFilePath := filepath.Join(tempDir, "tempFile.json")
	defer os.Remove(tempFilePath)

	// Write some text line to the file
	text := []byte(`blag`)
	err := os.WriteFile(tempFilePath, text, 0644)
	if err != nil {
		t.Fatalf("Failed to write to temporary file: %s", err)
	}

	// Use the NewIndexFile function to load the data from the file
	indexFile, err := NewCacheIndex(tempFilePath)

	// Assert that the data was loaded correctly
	assert.Nil(t, indexFile)
	assert.NotNil(t, err)
}
