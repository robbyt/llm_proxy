package storage

type StorageLayer interface {
	// GetBytes gets a value from the database using a byte key
	GetBytes(identifier string, key []byte) ([]byte, error)
	// GetBytesSafe attempts to get a value from the database, and returns nil if not found
	GetBytesSafe(identifier string, key []byte) ([]byte, error)
	// GetStr gets a value from the database using a string key
	GetStr(identifier string, key string) ([]byte, error)
	// GetStrSafe attempts to get a value from the database, and if it fails it returns nil
	GetStrSafe(identifier string, key string) ([]byte, error)
	// SetBytes sets a value in the database using a byte key
	SetBytes(identifier string, key, value []byte) error
	// SetStr sets a value in the database using a string key
	SetStr(identifier string, key, value string) error
	// Close closes the database
	Close() error
	// GetDBFileName returns the on-disk filename of the database
	GetDBFileName() string
}
