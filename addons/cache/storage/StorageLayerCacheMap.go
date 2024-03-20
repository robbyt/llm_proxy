package storage

import "sync"

// StorageLayerCacheMap is a collection of StorageLayer objects currently loaded into RAM
type StorageLayerCacheMap struct {
	sync.Mutex
	m map[string]StorageLayer
}

// Put adds a new StorageLayer object to this in-memory collection
func (sm *StorageLayerCacheMap) Put(identifier string, db StorageLayer) {
	sm.Lock()
	defer sm.Unlock()
	sm.m[identifier] = db
}

// Get retrieves a StorageLayer from the collection, and a boolean indicating if it was found
func (sm *StorageLayerCacheMap) Get(identifier string) (StorageLayer, bool) {
	sm.Lock()
	defer sm.Unlock()
	value, exists := sm.m[identifier]
	return value, exists
}

// All retrieves all StorageLayer objects from the collection
func (sm *StorageLayerCacheMap) All() map[string]StorageLayer {
	sm.Lock()
	defer sm.Unlock()
	return sm.m
}

// Delete removes a StorageLayer object from the cache collection
func (sm *StorageLayerCacheMap) Delete(identifier string) {
	sm.Lock()
	defer sm.Unlock()
	delete(sm.m, identifier)
}

// Len returns the number of StorageLayer objects in the cache collection
func (sm *StorageLayerCacheMap) Len() int {
	sm.Lock()
	defer sm.Unlock()
	return len(sm.m)
}

// Clear removes all StorageLayer objects from the cache collection
func (sm *StorageLayerCacheMap) Clear() {
	sm.Lock()
	defer sm.Unlock()
	sm.m = make(map[string]StorageLayer)
}

// NewStorageLayerCacheMap returns a new StorageLayerCacheMap object
func NewStorageLayerCacheMap() *StorageLayerCacheMap {
	return &StorageLayerCacheMap{
		m: make(map[string]StorageLayer),
	}
}
