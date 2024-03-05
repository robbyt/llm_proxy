package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/robbyt/llm_proxy/addons/fileUtils"
	log "github.com/sirupsen/logrus"
)

const (
	currentCacheConfigVer    = "v1"
	cacheConfigFileName      = "llm_proxy_cache.json"
	currentStorageVersion    = "v1"
	defaultStorageEngineName = "badger"
)

// cacheBehavior stores input args config for the cache
type cacheBehavior struct {
	Dir string // Directory to store the cache files
	TTL int64  // Time to live for cache files in seconds (0 means cache forever)
}

// CacheConfig is a struct that backs a llm_proxy_cache.json file, which configures the cache storage object
type CacheConfig struct {
	filePath       string `json:"-"`               // The full path of this cache index json file.
	ConfigVersion  string `json:"config_version"`  // The schema version of this cache index file.
	StorageEngine  string `json:"storage_engine"`  // The storage engine used for this cache
	StorageVersion string `json:"storage_version"` // The storage version used for this cache
	StoragePath    string `json:"storage_path"`    // The full path to the storage bucket (file path or database URI)
}

// GetStorageEngine returns the storage engine for the cache index file
/*
func (i *CacheConfig) GetStorageEngine() (storage.CacheDB, error) {
	// TODO get this out of the config layer
	return i.CacheStorageDB.GetStorageEngine()
}
*/

// Save writes the cache config json file to disk
func (i CacheConfig) Save() error {
	// Ensure the storage path subdirectory exists
	if err := os.MkdirAll(filepath.Dir(i.StoragePath), 0700); err != nil {
		return err
	}

	// Set the schema version if it's not already set
	if i.ConfigVersion == "" {
		i.ConfigVersion = currentCacheConfigVer
	}

	// Convert the IndexFile object to a JSON string
	jsonData, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON string to a tmp file, then rename it to the final file path
	tmpFile, err := os.CreateTemp(filepath.Dir(i.filePath), "llm_proxy_cache.json")
	if err != nil {
		return err
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	if err = os.WriteFile(tmpFile.Name(), jsonData, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile.Name(), i.filePath)
}

// Load reads the cache config json file from disk
func (i *CacheConfig) Load() error {
	existingFilePath := i.filePath
	jsonData, err := os.ReadFile(existingFilePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(jsonData, i); err != nil {
		return err
	}

	i.filePath = existingFilePath
	return nil
}

// NewCacheConfig creates a new IndexFile object to help with loading/saving meta-state as a json file.
// This object's purpose is to help loading the other database objects by pointing to their
// connection settings or file paths.
//
// cacheDir: the directory where the cache index file will be stored
func NewCacheConfig(cacheDir string) (*CacheConfig, error) {
	indexFilePath := filepath.Join(cacheDir, cacheConfigFileName)
	iFile := &CacheConfig{
		filePath:       indexFilePath,
		ConfigVersion:  currentCacheConfigVer,
		StorageEngine:  defaultStorageEngineName,
		StorageVersion: currentStorageVersion,
		StoragePath:    filepath.Join(cacheDir, "cache"),
	}

	if fileUtils.FileExists(iFile.filePath) {
		log.Debugf("Loading existing cache config file from: %s", iFile.filePath)
		if err := iFile.Load(); err != nil {
			return nil, fmt.Errorf("failed to load cache config file: %s", err)
		}
		return iFile, nil
	}

	log.Debugf("Creating a new cache config file at: %s", iFile.filePath)
	err := iFile.Save()
	if err != nil {
		return nil, fmt.Errorf("failed to create config file: %s", err)
	}
	return iFile, nil
}
