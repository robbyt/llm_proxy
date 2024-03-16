package storage

import (
	"fmt"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/robbyt/llm_proxy/addons/fileUtils"
)

type BadgerDB struct {
	DB         *badger.DB
	identifier string
	dbFile     string
}

func (b *BadgerDB) Close() error {
	return b.DB.Close()
}

// NewBadgerDB creates a wrapper object for a BadgerDB database
// identifier: identify the database (probably the request URL)
// dbFileDir: the directory where the database file is or will be stored
func NewBadgerDB(identifier string, cacheDir string) (*BadgerDB, error) {
	dbFile := fileUtils.ConvertURLtoFileName(cacheDir, identifier)

	db, err := badger.Open(badger.DefaultOptions(dbFile))
	if err != nil {
		return nil, fmt.Errorf("error opening db: %s", err)
	}
	return &BadgerDB{
		DB:         db,
		identifier: identifier,
		dbFile:     dbFile,
	}, nil
}
