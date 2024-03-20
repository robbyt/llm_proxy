package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"

	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"

	"github.com/robbyt/llm_proxy/addons/cache/storage/boltDB_Engine"
)

// BoltMetaDB is a single boltDB with multiple internal "buckets" for each URL (like tables)
type BoltMetaDB struct {
	dbFileDir string            // several DBs stored in the same directory, one for each base URL
	db        *boltDB_Engine.DB // the main db struct
	once      sync.Once
}

// Close closes all the BadgerDBs in the collection
func (c *BoltMetaDB) Close() error {
	var err error
	c.once.Do(func() {
		err = c.db.Close()
	})
	return err
}

// Get receives a request, pulls out the request URL, uses that URL as a
// cache "identifier" (to use the correct storage DB), and then looks up the
// request in cache based on the body, returning the cached response if found.
//
// The request URL can be considered the primary index (different files per URL),
// and the body is the secondary index.
func (c *BoltMetaDB) Get(req px.Request) (*px.Response, error) {
	if req.URL == nil || req.URL.String() == "" {
		return nil, fmt.Errorf("request URL is nil or empty")
	}

	identifier := req.URL.String()
	body := req.Body

	valueBytes, err := c.db.GetBytesSafe(identifier, body)
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

// Put receives a request and response, pulls out the request URL, uses that
// URL as a cache "identifier" (to use the correct storage DB), and then stores
// the response in cache based on the request body.
func (c *BoltMetaDB) Put(req px.Request, resp *px.Response) error {
	if req.URL == nil || req.URL.String() == "" {
		return fmt.Errorf("request URL is nil or empty")
	}
	identifier := req.URL.String()
	body := req.Body
	if len(body) == 0 {
		// if the body is empty, use a single space as the key because badger doesn't support empty keys
		body = []byte(" ")
	}

	// Encode the response into a gob object, for storage in the targetDB
	var reqBuffer bytes.Buffer
	enc := gob.NewEncoder(&reqBuffer)
	err := enc.Encode(req) // encode the request into a buffer object
	if err != nil {
		log.Fatal("encode error:", err)
	}

	// Store the encoded data in the targetDB
	err = c.db.SetBytes(identifier, body, reqBuffer.Bytes())
	if err != nil {
		log.Fatal("set bytes error:", err)
	}

	log.Debugf("stored response in cache for: %s", identifier)
	return nil
}

// NewBoltMetaDB creates a new BoltMetaDB object, to load or create a new boltDB on disk
func NewBoltMetaDB(dbFileDir string) (*BoltMetaDB, error) {
	bMeta := &BoltMetaDB{
		dbFileDir: dbFileDir,
	}
	return bMeta, nil
}
