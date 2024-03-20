package storage

import (
	"fmt"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

// BoltDB is a wrapper for the BoltDB database library
type BoltDB struct {
	db   *bolt.DB
	once sync.Once
}

// Close closes the database and runs other cleanup tasks
func (b *BoltDB) Close() error {
	var err error
	b.once.Do(func() {
		errClose := b.db.Close()
		if errClose != nil {
			err = fmt.Errorf("error closing db: %s", errClose)
			return
		}
	})
	return err
}

func (b *BoltDB) GetDBFileName() string {
	return b.db.Path()
}

func configBolt() *bolt.Options {
	return &bolt.Options{
		Timeout: 1 * time.Second,
	}
}

// NewBoltDB creates a wrapper object for a NewBoltDB database to creates new or load an existing DB.
// dbFileName: the path where the BoltDB file is stored on disk
func NewBoltDB(dbFileName string) (*BoltDB, error) {
	db, err := bolt.Open(dbFileName, 0600, configBolt())
	if err != nil {
		return nil, fmt.Errorf("error opening db: %s", err)
	}
	return &BoltDB{db: db}, nil
}
