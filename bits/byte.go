package bits

// IsEmpty checks if the input slice of bytes is empty.
func IsEmpty(input []byte) bool {
	return len(input) == 0
}
