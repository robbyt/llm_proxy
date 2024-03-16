package cache

import "fmt"

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
