package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"path/filepath"
	"sync"

	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"

	"github.com/robbyt/llm_proxy/addons/cache/storage"
	"github.com/robbyt/llm_proxy/addons/fileUtils"
)

const (
	bdbControlDbName = "control"
)

// BadgerMetaDB is a collection of multiple BadgerDBs, each with a different identifier
type BadgerMetaDB struct {
	dbFileDir         string            // several DBs stored in the same directory, one for each base URL
	controlDb         *storage.BadgerDB // the control DB, stores the map of URL -> dbFilePath
	controlDBAddMutex sync.Mutex
	dbCache           *storage.BadgerDB_CacheMap
	metaIsClosed      bool
}

// controlDB_FilePath returns the filepath to the control DB, which handles lookup of the other DBs
func (c *BadgerMetaDB) controlDB_FilePath() string {
	return filepath.Join(c.dbFileDir, bdbControlDbName)
}

// filePath returns the file path for a given identifier (i.e., URL)
func (c *BadgerMetaDB) filePath(identifier string) string {
	return fileUtils.ConvertIDtoFileName(c.dbFileDir, identifier)
}

// addDB adds a new BadgerDB to the meta collection
func (c *BadgerMetaDB) addDB(identifier string) error {
	c.controlDBAddMutex.Lock()
	defer c.controlDBAddMutex.Unlock()

	db, err := c.loadDB(identifier, c.filePath(identifier))
	if err != nil {
		return err
	}
	return c.controlDb.SetStr(identifier, db.DBFileName)
}

// getDB retrieves a BadgerDB from the collection
func (c *BadgerMetaDB) getDB(identifier string) (*storage.BadgerDB, error) {
	dbFileName, err := c.controlDb.GetStrSafe(identifier)
	if err != nil {
		log.Warnf("error getting dbFileName from control db for %s: %s", identifier, err)
		return nil, err
	}

	if dbFileName == nil {
		log.Debugf("dbFileName not found in control db, creating: %s", identifier)
		if err := c.addDB(identifier); err != nil {
			return nil, err
		}

		// try again, this time without the safe lookup
		dbFileName, err = c.controlDb.GetStr(identifier)
		if err != nil {
			return nil, err
		}
	}

	return c.loadDB(identifier, string(dbFileName))
}

// loadDB loads or creates a badgerDB to/from disk
func (c *BadgerMetaDB) loadDB(identifier, dbFileName string) (*storage.BadgerDB, error) {
	// lookup the db in the cache
	if db, ok := c.dbCache.Get(identifier); ok {
		return db, nil
	}

	bDb, err := storage.NewBadgerDB(identifier, dbFileName)
	if err != nil {
		return nil, err
	}

	c.dbCache.Put(identifier, bDb)
	return bDb, nil
}

/*
func (c *BadgerMetaDB) closeSerial() error {
	errors := make([]error, 0)
	for _, db := range c.dbCache.All() {
		if err := db.Close(); err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("error(s) closing db: %v", errors)
	}

	return c.controlDb.Close()
}
*/

func (c *BadgerMetaDB) closeParallel() error {
	var wg sync.WaitGroup
	errors := make(chan error)
	defer c.controlDb.Close()

	for _, db := range c.dbCache.All() {
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

// Close closes all the BadgerDBs in the collection
func (c *BadgerMetaDB) Close() error {
	if c.metaIsClosed {
		log.Warn("attempted to close an already closed BadgerMetaDB")
		return nil
	}

	c.metaIsClosed = true
	return c.closeParallel()
}

// Lookup receives a request, pulls out the request URL, uses that URL as a
// cache "identifier" (to use the correct storage DB), and then looks up the
// request in cache based on the body, returning the cached response if found.
//
// The request URL can be considered the primary index (different files per URL),
// and the body is the secondary index.
func (c *BadgerMetaDB) Lookup(req px.Request) (*px.Response, error) {
	if req.URL == nil || req.URL.String() == "" {
		return nil, fmt.Errorf("request URL is nil or empty")
	}
	identifier := req.URL.String()

	body := req.Body
	if len(body) == 0 {
		// if the body is empty, use a single space as the key because badger doesn't support empty keys
		body = []byte(" ")
	}

	targetDB, err := c.getDB(identifier)
	if err != nil {
		return nil, err
	}
	if targetDB == nil {
		return nil, fmt.Errorf("targetDB is nil for identifier: %s", identifier)
	}

	valueBytes, err := targetDB.GetBytesSafe(body)
	if err != nil {
		return nil, err
	}
	if valueBytes == nil {
		log.Debugf("valueBytes empty for: %s", identifier)
		return nil, nil
	}

	decoder := gob.NewDecoder(bytes.NewReader(valueBytes))
	var r px.Response // damn, maybe this won't work?
	if err := decoder.Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}

func (c *BadgerMetaDB) Store(req px.Request, resp *px.Response) error {
	if req.URL == nil || req.URL.String() == "" {
		return fmt.Errorf("request URL is nil or empty")
	}
	identifier := req.URL.String()
	body := req.Body
	if len(body) == 0 {
		// if the body is empty, use a single space as the key because badger doesn't support empty keys
		body = []byte(" ")
	}
	targetDB, err := c.getDB(identifier)
	if err != nil {
		return err
	}
	if targetDB == nil {
		return fmt.Errorf("targetDB is nil for identifier: %s", identifier)
	}

	// Encode the response into a gob object, for storage in the targetDB
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(body)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	// Store the encoded data in the targetDB
	err = targetDB.SetStr(identifier, buf.String())
	if err != nil {
		log.Fatal("set bytes error:", err)
	}

	log.Debugf("stored response in cache for: %s", identifier)
	return nil
}

func NewBadgerMetaDB(dbFileDir string) (*BadgerMetaDB, error) {
	bMeta := &BadgerMetaDB{
		dbFileDir: dbFileDir,
		dbCache:   storage.NewBadgerDB_CacheMap(),
	}

	// create/load the control db, which does mapping from URL -> bdb object
	controlDB, err := bMeta.loadDB(bdbControlDbName, bMeta.controlDB_FilePath())
	if err != nil {
		return nil, err
	}
	bMeta.controlDb = controlDB

	return bMeta, nil
}
