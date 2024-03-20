package boltDB_Engine

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

// deleteForKey is a private/internal method to delete a key from the database w/o formatting
func (b *DB) deleteForKey(identifier string, key []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(identifier))
		if err != nil {
			return fmt.Errorf("error creating/loading bucket: %s", err)
		}

		return bucket.Delete([]byte(key))
	})
}

// Delete removes a key from the database
func (b *DB) Delete(identifier string, key []byte) error {
	return b.deleteForKey(identifier, keyFormatter(key))
}
