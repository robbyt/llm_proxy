package storage

import (
	"fmt"
	"sync"

	badger "github.com/dgraph-io/badger/v4"
	log "github.com/sirupsen/logrus"
)

// local variable to set the log level for the BadgerDB output
var bDb_LogLevel = log.DebugLevel

type BadgerDB struct {
	DB         *badger.DB
	once       sync.Once
	gcRunning  sync.Mutex
	identifier string
	DBFileName string
}

func (b *BadgerDB) GetBytes(key []byte) ([]byte, error) {
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

func (b *BadgerDB) GetStr(key string) ([]byte, error) {
	return b.GetBytes([]byte(key))
}

// GetStrSafe attempts to get a value from the database, and if it fails it returns nil
func (b *BadgerDB) GetStrSafe(key string) ([]byte, error) {
	val, err := b.GetStr(key)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting value from db: %s", err)
	}
	return val, nil
}

func (b *BadgerDB) SetBytes(key, value []byte) error {
	// badger.NewEntry(key, value).WithMeta(byte(42))

	return b.DB.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (b *BadgerDB) SetStr(key, value string) error {
	return b.SetBytes([]byte(key), []byte(value))
}

func (b *BadgerDB) RunGC() error {
	b.gcRunning.Lock()
	defer b.gcRunning.Unlock()
	log.Debug("Running BadgerDB GC")
	return b.DB.RunValueLogGC(0.5)
}

func (b *BadgerDB) Close() error {
	var err error
	b.RunGC()

	b.once.Do(func() {
		// log.Debug("Closing BadgerDB")
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

// NewBadgerDB creates a wrapper object for a BadgerDB database
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
