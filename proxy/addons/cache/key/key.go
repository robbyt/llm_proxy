package key

// NewKey creates a new Key object using the default hash algorithm
func NewKey(key []byte) Key {
	return NewBLAKE2Key(key)
}

// NewKeyStr creates a new Key object from a string using the default hash algorithm
func NewKeyStr(key string) Key {
	return NewBLAKE2Key([]byte(key))
}
