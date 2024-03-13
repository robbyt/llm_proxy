package storage

import (
	"fmt"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/robbyt/llm_proxy/addons/fileUtils"
)

type bDB struct {
	db         *badger.DB
	identifier string
	dbFile     string
}

// new_bDB creates a wrapper object for a BadgerDB database
// url: the base URL for requests stored in this DB
// dbFileDir: the directory where the database file will be stored
func new_bDB(identifier string, dbFileDir string) (*bDB, error) {
	dbFile := fileUtils.ConvertURLtoFileName(dbFileDir, identifier)
	err := fileUtils.RelocateExistingFileIfExists(dbFile)
	if err != nil {
		return nil, fmt.Errorf("error while relocating existing file: %s", err)
	}

	db, err := badger.Open(badger.DefaultOptions(dbFile))
	if err != nil {
		return nil, fmt.Errorf("error opening db: %s", err)
	}
	return &bDB{
		db:         db,
		identifier: identifier,
		dbFile:     dbFile,
	}, nil
}

func (b *bDB) close() error {
	return b.db.Close()
}
