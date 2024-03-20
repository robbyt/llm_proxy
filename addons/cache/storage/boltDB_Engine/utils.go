package boltDB_Engine

// keyFormatter is a helper function to format a key for storage
func keyFormatter(key []byte) []byte {
	if len(key) == 0 {
		return []byte("nil")
	}
	return key
}
