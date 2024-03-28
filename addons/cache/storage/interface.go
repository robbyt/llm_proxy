package storage

// Engine is an interface for a storage engine, backed by some database for persistence
type Engine interface {
	// Get gets a value from the database using a byte key
	GetBytes(identifier string, key []byte) ([]byte, error)
	// GetBytesSafe attempts to get a value from the database, and returns nil if not found
	GetBytesSafe(identifier string, key []byte) ([]byte, error)
	// SetBytes sets a value in the database using a byte key
	SetBytes(identifier string, key, value []byte) error
	// Close closes the database
	Close() error
	// GetDBFileName returns the on-disk filename of the database
	GetDBFileName() string
}
