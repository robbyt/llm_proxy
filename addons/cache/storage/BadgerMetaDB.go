package storage

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"

	badger "github.com/dgraph-io/badger/v4"
	px "github.com/kardianos/mitmproxy/proxy"
)

// BadgerMetaDB is a collection of multiple BadgerDBs, each with a different identifier
type BadgerMetaDB struct {
	db        map[string]*bDB // key is the base URL
	dbFileDir string          // several DBs stored in the same directory, one for each base URL
}

// addDB adds a new BadgerDB to the meta collection
func (c *BadgerMetaDB) addDB(identifier string) error {
	_, err := c.getDB(identifier)
	if err != nil {
		return fmt.Errorf("db already exists for identifier: %s", identifier)
	}

	db, err := new_bDB(identifier, c.dbFileDir)
	if err != nil {
		return err
	}

	c.db[identifier] = db
	return nil
}

// getDB retrieves a BadgerDB from the collection
func (c *BadgerMetaDB) getDB(identifier string) (*bDB, error) {
	targetDB := c.db[identifier]
	if targetDB == nil {
		return nil, fmt.Errorf("no db found in meta for identifier: %s", identifier)
	}
	return targetDB, nil
}

// Close closes all the BadgerDBs in the collection
func (c *BadgerMetaDB) Close() error {
	return c.closeSerial()
}

func (c *BadgerMetaDB) closeSerial() error {
	errors := make([]error, 0)
	for _, db := range c.db {
		err := db.close()
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("error(s) closing db: %v", errors)
	}

	return nil
}

func (c *BadgerMetaDB) closeParallel() error {
	var wg sync.WaitGroup
	errors := make(chan error)

	for _, db := range c.db {
		wg.Add(1)
		go func(db *bDB) {
			defer wg.Done()
			if err := db.close(); err != nil {
				errors <- err
			}
		}(db)
	}

	// Close the errors channel after all goroutines have finished
	go func() {
		wg.Wait()
		close(errors)
	}()

	// Collect all the errors
	errs := make([]error, 0)
	for err := range errors {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("error(s) closing db: %v", errs)
	}

	return nil
}

// GetBytes retrieves a value from one of the BadgerDBs in the collection
func (c *BadgerMetaDB) GetBytes(identifier string, key []byte) ([]byte, error) {
	targetDB, err := c.getDB(identifier)
	if err != nil {
		return nil, err
	}

	var value []byte
	err = targetDB.db.View(func(txn *badger.Txn) error {
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

func (c *BadgerMetaDB) GetStr(identifier string, key string) ([]byte, error) {
	return c.GetBytes(identifier, []byte(key))
}

func (c *BadgerMetaDB) SetBytes(identifier string, key, value []byte) error {
	badger.NewEntry(key, value).WithMeta(byte(42))

	targetDB := c.db[identifier]
	if targetDB == nil {
		return fmt.Errorf("no db found for url: %s", identifier)
	}

	return targetDB.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (c *BadgerMetaDB) SetStr(identifier, key string, value []byte) error {
	return c.SetBytes(identifier, []byte(key), value)
}

func (c *BadgerMetaDB) Lookup(req px.Request) (*px.Response, error) {
	if req.URL == nil || req.URL.String() == "" {
		return nil, fmt.Errorf("request URL is nil or empty")
	}
	identifier := req.URL.String()

	var body []byte
	if req.Body != nil {
		body = req.Body
	} else {
		body = []byte("")
	}

	valueBytes, err := c.GetBytes(identifier, body)
	if err != nil {
		return nil, err
	}

	decoder := gob.NewDecoder(bytes.NewReader(valueBytes))
	var r px.Response // damn, this won't work
	if err := decoder.Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}

func NewBadgerDB(dbFileDir string) (*BadgerMetaDB, error) {
	return &BadgerMetaDB{
		db:        make(map[string]*bDB),
		dbFileDir: dbFileDir,
	}, nil
}
