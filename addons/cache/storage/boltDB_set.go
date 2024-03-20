package storage

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

// setBytesForKey is a private/internal method to set a value in the database w/o formatting
func (b *BoltDB) setBytesForKey(identifier string, key, value []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(identifier))
		if err != nil {
			return fmt.Errorf("error creating/loading bucket: %s", err)
		}

		err = bucket.Put(key, value)
		if err != nil {
			return fmt.Errorf("error putting value: %s", err)
		}
		return nil
	})
}

// SetBytes sets a value in the database using a byte key
func (b *BoltDB) SetBytes(identifier string, key, value []byte) error {
	return b.setBytesForKey(identifier, keyFormatter(key), value)
}

// SetStr sets a value in the database using a string key
func (b *BoltDB) SetStr(identifier string, key, value string) error {
	return b.SetBytes(identifier, []byte(key), []byte(value))
}
