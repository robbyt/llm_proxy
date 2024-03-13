package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/robbyt/llm_proxy/addons/cache/storage"
	"github.com/robbyt/llm_proxy/addons/fileUtils"
)

const (
	latestIndexVersion   = "v1"
	latestStorageVersion = "v1"
	IndexFileName        = "cache_index.json"
)

type CacheStorage struct {
	Name           string `json:"name"`            // The name of the storage bucket (human-readable name or database name)
	Path           string `json:"path"`            // The full path to the storage bucket (file path or database URI)
	StorageEngine  string `json:"storage_engine"`  // The storage engine used for this cache
	StorageVersion string `json:"storage_version"` // The storage version used for this cache
}

func (csb *CacheStorage) GetStorageEngine() (storage.CacheDB, error) {
	switch csb.StorageEngine {
	case "badger":
		return storage.NewBadgerDB(csb.Path)
	default:
		return nil, fmt.Errorf("unknown storage engine: %s", csb.StorageEngine)
	}
}

// CacheIndexFile is a struct that backs a cache_index.json file
type CacheIndexFile struct {
	FilePath      string       `json:"-"`              // The full path of this cache index file.
	SchemaVersion string       `json:"schema_version"` // The schema version of this cache index file.
	CacheStorage  CacheStorage `json:"cache_storage"`  // The storage buckets used for this cache
}

// Save writes the index file to disk
func (i CacheIndexFile) Save() error {
	// Set the schema version if it's not already set
	if i.SchemaVersion == "" {
		i.SchemaVersion = latestIndexVersion
	}

	// Convert the IndexFile object to a JSON string
	jsonData, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON string to a file
	err = os.WriteFile(i.FilePath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Load reads the index file from disk
func (i *CacheIndexFile) Load() error {
	// Read the index file from disk
	existingFilePath := i.FilePath
	jsonData, err := os.ReadFile(existingFilePath)
	if err != nil {
		return err
	}

	// Convert the JSON string to an IndexFile object
	err = json.Unmarshal(jsonData, i)
	i.FilePath = existingFilePath
	return err
}

// GetStorageEngine returns the storage engine for the cache index file
func (i *CacheIndexFile) GetStorageEngine() (storage.CacheDB, error) {
	return i.CacheStorage.GetStorageEngine()
}

// NewCacheIndex creates a new IndexFile object, which can be used to load/save the file to disk
func NewCacheIndex(cacheDir string) (*CacheIndexFile, error) {
	indexFilePath := filepath.Join(cacheDir, IndexFileName)
	iFile := &CacheIndexFile{
		FilePath: indexFilePath,
	}

	if fileUtils.FileExists(indexFilePath) {
		err := iFile.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load index file: %s", err)
		}
	}

	return iFile, nil
}
