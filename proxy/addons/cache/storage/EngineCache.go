package storage

import "sync"

// EngineCache is a collection of storage engine objects currently loaded into RAM
type EngineCache struct {
	sync.Mutex
	m map[string]Engine
}

// Put adds a new StorageLayer object to this in-memory collection
func (sm *EngineCache) Put(identifier string, db Engine) {
	sm.Lock()
	defer sm.Unlock()
	sm.m[identifier] = db
}

// Get retrieves a StorageLayer from the collection, and a boolean indicating if it was found
func (sm *EngineCache) Get(identifier string) (Engine, bool) {
	sm.Lock()
	defer sm.Unlock()
	value, exists := sm.m[identifier]
	return value, exists
}

// All retrieves all StorageLayer objects from the collection
func (sm *EngineCache) All() map[string]Engine {
	sm.Lock()
	defer sm.Unlock()
	return sm.m
}

// Delete removes a StorageLayer object from the cache collection
func (sm *EngineCache) Delete(identifier string) {
	sm.Lock()
	defer sm.Unlock()
	delete(sm.m, identifier)
}

// Len returns the number of StorageLayer objects in the cache collection
func (sm *EngineCache) Len() int {
	sm.Lock()
	defer sm.Unlock()
	return len(sm.m)
}

// Clear removes all StorageLayer objects from the cache collection
func (sm *EngineCache) Clear() {
	sm.Lock()
	defer sm.Unlock()
	sm.m = make(map[string]Engine)
}

// NewCacheMap returns a new StorageLayerCacheMap object
func NewCacheMap() *EngineCache {
	return &EngineCache{
		m: make(map[string]Engine),
	}
}
