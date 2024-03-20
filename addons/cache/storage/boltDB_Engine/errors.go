package boltDB_Engine

import "fmt"

// BucketNotFoundError is returned when a bucket/table is not found
type BucketNotFoundError struct {
	Identifier string
}

func (e BucketNotFoundError) Error() string {
	return fmt.Sprintf("Bucket not found: %s", e.Identifier)
}

// DatabaseExistsError is returned when a database already exists when trying to create it
type ErrKeyNotFound struct {
	Identifier string
	Key        []byte
}

func (e ErrKeyNotFound) Error() string {
	return fmt.Sprintf("not found! identifier: %s key: %s ", e.Identifier, e.Key)
}
