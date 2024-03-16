package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"

	badger "github.com/dgraph-io/badger/v4"
	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"

	"github.com/robbyt/llm_proxy/addons/cache/storage"
)

const (
	bdbControlDbName = "system"
)

// BadgerMetaDB is a collection of multiple BadgerDBs, each with a different identifier
type BadgerMetaDB struct {
	db        map[string]*storage.BadgerDB // key is the base URL
	dbFileDir string                       // several DBs stored in the same directory, one for each base URL
}

// addDB adds a new BadgerDB to the meta collection
func (c *BadgerMetaDB) addDB(identifier string) error {
	bDB, _ := c.getDB(identifier)
	if bDB != nil {
		return DatabaseExistsError{Identifier: identifier}
	}

	db, err := storage.NewBadgerDB(identifier, c.dbFileDir)
	if err != nil {
		return err
	}

	c.db[identifier] = db
	return nil
}

// getDB retrieves a BadgerDB from the collection
func (c *BadgerMetaDB) getDB(identifier string) (*storage.BadgerDB, error) {
	targetDB := c.db[identifier]
	if targetDB == nil {
		return nil, DatabaseNotFoundError{Identifier: identifier}
	}
	return targetDB, nil
}

func (c *BadgerMetaDB) closeSerial() error {
	errors := make([]error, 0)
	for _, db := range c.db {
		if err := db.Close(); err != nil {
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
		go func(db *storage.BadgerDB) {
			defer wg.Done()
			if err := db.Close(); err != nil {
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

// getBytes retrieves a value from one of the BadgerDBs in the collection.
// Not yet sure how to handle key not found errors.
func (c *BadgerMetaDB) getBytes(identifier string, key []byte) ([]byte, error) {
	targetDB, err := c.getDB(identifier)
	if err != nil {
		return nil, err
	}

	var value []byte
	err = targetDB.DB.View(func(txn *badger.Txn) error {
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

func (c *BadgerMetaDB) getStr(identifier string, key string) ([]byte, error) {
	return c.getBytes(identifier, []byte(key))
}

func (c *BadgerMetaDB) setBytes(identifier string, key, value []byte) error {
	badger.NewEntry(key, value).WithMeta(byte(42))

	targetDB := c.db[identifier]
	if targetDB == nil {
		return fmt.Errorf("no db found for url: %s", identifier)
	}

	return targetDB.DB.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (c *BadgerMetaDB) SetStr(identifier, key string, value []byte) error {
	return c.setBytes(identifier, []byte(key), value)
}

// Close closes all the BadgerDBs in the collection
func (c *BadgerMetaDB) Close() error {
	return c.closeSerial()
}

// Lookup receives a request, pulls out the request URL, uses that URL as a
// cache "identifier" (to use the correct DB), and then looks up the request
// in the cache based on the body, returning the response if found.
func (c *BadgerMetaDB) Lookup(req px.Request) (*px.Response, error) {
	if req.URL == nil || req.URL.String() == "" {
		return nil, fmt.Errorf("request URL is nil or empty")
	}
	identifier := req.URL.String()

	body := req.Body
	if body == nil {
		body = []byte("")
	}

	// unsure if err means not found, or some other error
	valueBytes, err := c.getBytes(identifier, body)
	if err != nil {
		log.Warnf("error looking up cache: %s", err)
		return nil, err
	}

	decoder := gob.NewDecoder(bytes.NewReader(valueBytes))
	var r px.Response // damn, this won't work?
	if err := decoder.Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}

func (c *BadgerMetaDB) Store(req px.Request, resp *px.Response) error {
	log.Errorf("Store not implemented")
	return nil
}

func NewBadgerDB(dbFileDir string) (*BadgerMetaDB, error) {
	bMeta := &BadgerMetaDB{
		db:        make(map[string]*storage.BadgerDB),
		dbFileDir: dbFileDir,
	}

	err := bMeta.addDB(bdbControlDbName)
	if err != nil {
		// ignore if the control DB already exists
		if _, ok := err.(*DatabaseExistsError); !ok {
			return nil, err
		}
	}

	return bMeta, nil
}
