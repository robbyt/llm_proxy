package storage

import "sync"

// CacheMap is a collection of StorageLayer objects currently loaded into RAM
type CacheMap struct {
	sync.Mutex
	m map[string]StorageLayer
}

// Put adds a new StorageLayer object to this in-memory collection
func (sm *CacheMap) Put(identifier string, db StorageLayer) {
	sm.Lock()
	defer sm.Unlock()
	sm.m[identifier] = db
}

// Get retrieves a StorageLayer from the collection, and a boolean indicating if it was found
func (sm *CacheMap) Get(identifier string) (StorageLayer, bool) {
	sm.Lock()
	defer sm.Unlock()
	value, exists := sm.m[identifier]
	return value, exists
}

// All retrieves all StorageLayer objects from the collection
func (sm *CacheMap) All() map[string]StorageLayer {
	sm.Lock()
	defer sm.Unlock()
	return sm.m
}

// Delete removes a StorageLayer object from the cache collection
func (sm *CacheMap) Delete(identifier string) {
	sm.Lock()
	defer sm.Unlock()
	delete(sm.m, identifier)
}

// Len returns the number of StorageLayer objects in the cache collection
func (sm *CacheMap) Len() int {
	sm.Lock()
	defer sm.Unlock()
	return len(sm.m)
}

// Clear removes all StorageLayer objects from the cache collection
func (sm *CacheMap) Clear() {
	sm.Lock()
	defer sm.Unlock()
	sm.m = make(map[string]StorageLayer)
}

// NewCacheMap returns a new StorageLayerCacheMap object
func NewCacheMap() *CacheMap {
	return &CacheMap{
		m: make(map[string]StorageLayer),
	}
}
