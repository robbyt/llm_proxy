package storage

import "fmt"

// DatabaseExistsError is returned when a database already exists when trying to create it
type ErrKeyNotFound struct {
	Identifier string
	Key        []byte
}

func (e ErrKeyNotFound) Error() string {
	return fmt.Sprintf("not found! identifier: %s key: %s ", e.Identifier, e.Key)
}

// DatabaseExistsError is returned when a database already exists when trying to create it
type DatabaseExistsError struct {
	Identifier string
}

func (e DatabaseExistsError) Error() string {
	return fmt.Sprintf("Database already exists: %s", e.Identifier)
}

// DatabaseNotFoundError is returned when a database is not found
type DatabaseNotFoundError struct {
	Identifier string
}

func (e DatabaseNotFoundError) Error() string {
	return fmt.Sprintf("Database not found: %s", e.Identifier)
}

// BucketNotFoundError is returned when a bucket/table is not found
type BucketNotFoundError struct {
	Identifier string
}

func (e BucketNotFoundError) Error() string {
	return fmt.Sprintf("Bucket not found: %s", e.Identifier)
}
