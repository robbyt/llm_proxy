package bdb

import (
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

func (b *DB) getBytesForKey(identifier string, key []byte) (value []byte, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(identifier))
		if bucket == nil {
			return BucketNotFoundError{Identifier: identifier}
		}

		v := bucket.Get(key)
		if v == nil {
			return ErrKeyNotFound{Identifier: identifier, Key: key}
		}

		// Make a copy of the value, so it's readable outside of this transaction
		value = make([]byte, len(v))
		copy(value, v)

		return nil
	})
	return value, err
}

// GetBytes gets a value from the database using a byte key
func (b *DB) GetBytes(identifier string, key []byte) (value []byte, err error) {
	return b.getBytesForKey(identifier, keyFormatter(key))
}

// GetBytesSafe attempts to get a value from the database, and returns nil if not found
func (b *DB) GetBytesSafe(identifier string, key []byte) ([]byte, error) {
	val, err := b.GetBytes(identifier, key)
	if err != nil {
		var keyNotFoundError ErrKeyNotFound
		var bucketNotFoundError BucketNotFoundError
		if errors.As(err, &keyNotFoundError) || errors.As(err, &bucketNotFoundError) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting value: %s", err)
	}
	return val, nil
}

// GetStr gets a value from the database using a string key
func (b *DB) GetStr(identifier string, key string) ([]byte, error) {
	return b.GetBytes(identifier, []byte(key))
}

// GetStrSafe attempts to get a value from the database, and if it fails it returns nil
func (b *DB) GetStrSafe(identifier string, key string) ([]byte, error) {
	return b.GetBytesSafe(identifier, []byte(key))
}
