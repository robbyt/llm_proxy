package storage

import (
	"fmt"
	"sync"

	badger "github.com/dgraph-io/badger/v4"
	log "github.com/sirupsen/logrus"
)

// keyFormatter is a private/internal function to format keys
func keyFormatter(key []byte) []byte {
	if len(key) == 0 {
		return []byte("nil")
	}
	return key
}

// local variable to set the log level for the BadgerDB output
var bDb_LogLevel = log.DebugLevel

// BadgerDB is a wrapper for the BadgerDB database library
type BadgerDB struct {
	DB         *badger.DB
	once       sync.Once
	gcRunning  sync.Mutex
	identifier string
	DBFileName string
}

// deleteForKey is a private/internal method to delete a key from the database w/o formatting
func (b *BadgerDB) deleteForKey(key []byte) error {
	return b.DB.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Delete removes a key from the database
func (b *BadgerDB) Delete(key []byte) error {
	return b.deleteForKey(keyFormatter(key))
}

// getBytesForKey is a private/internal method to get a value from the database w/o formatting
func (b *BadgerDB) getBytesForKey(key []byte) ([]byte, error) {
	var value []byte
	err := b.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			value = val
			return nil
		})
		return err
	})
	return value, err
}

// GetBytes gets a value from the database using a byte key
func (b *BadgerDB) GetBytes(key []byte) ([]byte, error) {
	return b.getBytesForKey(keyFormatter(key))
}

// GetBytesSafe attempts to get a value from the database, and returns nil if not found
func (b *BadgerDB) GetBytesSafe(key []byte) ([]byte, error) {
	val, err := b.GetBytes(key)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting value from db: %s", err)
	}
	return val, nil
}

// GetStr gets a value from the database using a string key
func (b *BadgerDB) GetStr(key string) ([]byte, error) {
	return b.GetBytes([]byte(key))
}

// GetStrSafe attempts to get a value from the database, and if it fails it returns nil
func (b *BadgerDB) GetStrSafe(key string) ([]byte, error) {
	return b.GetBytesSafe([]byte(key))
}

// setBytesForKey is a private/internal method to set a value in the database w/o formatting
func (b *BadgerDB) setBytesForKey(key, value []byte) error {
	return b.DB.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

// SetBytes sets a value in the database using a byte key
func (b *BadgerDB) SetBytes(key, value []byte) error {
	return b.setBytesForKey(keyFormatter(key), value)
}

// SetStr sets a value in the database using a string key
func (b *BadgerDB) SetStr(key, value string) error {
	return b.SetBytes([]byte(key), []byte(value))
}

// RunGC runs the garbage collector on the database to remove stale data from disk
func (b *BadgerDB) RunGC() error {
	b.gcRunning.Lock()
	defer b.gcRunning.Unlock()
	log.Debug("Running BadgerDB GC")
	return b.DB.RunValueLogGC(0.5)
}

// Close closes the database and runs other cleanup tasks
func (b *BadgerDB) Close() error {
	var err error
	b.once.Do(func() {
		closeErr := b.RunGC()
		if closeErr != nil {
			log.Errorf("error running GC: %s", err)
		}

		errClose := b.DB.Close()
		if errClose != nil {
			err = fmt.Errorf("error closing db: %s", errClose)
			return
		}
	})
	return err
}

// configBadger tells the BadgerDB which DB to open sets other caching/sizing/scaling options
func configBadger(dbFileName string) badger.Options {
	options := badger.DefaultOptions(dbFileName).
		WithIndexCacheSize(1024 * 1024 * 20).
		WithNumVersionsToKeep(0).
		WithCompactL0OnClose(true).
		WithValueLogFileSize(1024 * 1024 * 100) // 100MB

	var logger = log.New()
	logger.SetLevel(bDb_LogLevel)
	options.Logger = logger

	return options
}

// NewBadgerDB creates a wrapper object for a BadgerDB database. This creates new or loads existing DBs.
// identifier: identify the database (probably the request URL)
// dbFileName: the full path to the database file is or will be stored
func NewBadgerDB(identifier, dbFileName string) (*BadgerDB, error) {
	db, err := badger.Open(configBadger(dbFileName))
	if err != nil {
		return nil, fmt.Errorf("error opening db: %s", err)
	}
	return &BadgerDB{
		DB:         db,
		identifier: identifier,
		DBFileName: dbFileName,
	}, nil
}
