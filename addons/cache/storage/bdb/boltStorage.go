package bdb

import (
	"fmt"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

// DB is a wrapper for the DB database library
type DB struct {
	db   *bolt.DB
	once sync.Once
}

// Close closes the database and runs other cleanup tasks
func (b *DB) Close() error {
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

func (b *DB) GetDBFileName() string {
	return b.db.Path()
}

func configBolt() *bolt.Options {
	return &bolt.Options{
		Timeout: 1 * time.Second,
	}
}

// NewBoltDB creates a wrapper object for a NewBoltDB database to creates new or load an existing DB.
// dbFileName: the path where the BoltDB file is stored on disk
func NewBoltDB(dbFileName string) (*DB, error) {
	db, err := bolt.Open(dbFileName, 0600, configBolt())
	if err != nil {
		return nil, fmt.Errorf("error opening db: %s", err)
	}
	return &DB{db: db}, nil
}
