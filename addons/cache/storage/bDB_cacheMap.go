package storage

import "sync"

// BadgerDB_CacheMap is a collection of BadgerDBs that are currently loaded into RAM
type BadgerDB_CacheMap struct {
	sync.Mutex
	m map[string]*BadgerDB
}

// Put adds a BadgerDB to the collection
func (sm *BadgerDB_CacheMap) Put(identifier string, bdb *BadgerDB) {
	sm.Lock()
	defer sm.Unlock()
	sm.m[identifier] = bdb
}

// Get retrieves a BadgerDB from the collection, and a boolean indicating if it was found
func (sm *BadgerDB_CacheMap) Get(identifier string) (*BadgerDB, bool) {
	sm.Lock()
	defer sm.Unlock()
	value, exists := sm.m[identifier]
	return value, exists
}

// All retrieves all BadgerDBs from the collection
func (sm *BadgerDB_CacheMap) All() map[string]*BadgerDB {
	sm.Lock()
	defer sm.Unlock()
	return sm.m
}

// Delete removes a BadgerDB pointer from the cache collection
func (sm *BadgerDB_CacheMap) Delete(identifier string) {
	sm.Lock()
	defer sm.Unlock()
	delete(sm.m, identifier)
}

// Len returns the number of BadgerDBs in the cache collection
func (sm *BadgerDB_CacheMap) Len() int {
	sm.Lock()
	defer sm.Unlock()
	return len(sm.m)
}

// Clear removes all BadgerDBs from the cache collection
func (sm *BadgerDB_CacheMap) Clear() {
	sm.Lock()
	defer sm.Unlock()
	sm.m = make(map[string]*BadgerDB)
}

// NewBadgerDB_CacheMap returns a new BadgerDB_CacheMap
func NewBadgerDB_CacheMap() *BadgerDB_CacheMap {
	return &BadgerDB_CacheMap{
		m: make(map[string]*BadgerDB),
	}
}
