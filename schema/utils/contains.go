package utils

// Contains checks if a slice contains a string.
func Contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
