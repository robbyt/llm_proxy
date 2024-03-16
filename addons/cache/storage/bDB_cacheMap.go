package storage

import "sync"

type BadgerDB_CacheMap struct {
	sync.RWMutex
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
	sm.RLock()
	defer sm.RUnlock()
	value, exists := sm.m[identifier]
	return value, exists
}

// All retrieves all BadgerDBs from the collection
func (sm *BadgerDB_CacheMap) All() map[string]*BadgerDB {
	sm.RLock()
	defer sm.RUnlock()
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
	sm.RLock()
	defer sm.RUnlock()
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
	bcm := &BadgerDB_CacheMap{}
	bcm.Clear()
	return bcm
}
