package gocache

// A read-only view of bytes stored in the cache.
type ByteView struct {
	bytes []byte
}

func (b ByteView) Len() int {
	return len(b.bytes)
}

// ByteSlice returns a copy of the data as a byte slice.
func (b ByteView) ByteSlice() []byte {
	return copyBytes(b.bytes)
}

// String returns the data as a string.
func (b ByteView) String() string {
	return string(b.bytes)
}

func copyBytes(bytes []byte) []byte {
	c := make([]byte, len(bytes))
	copy(c, bytes)
	return c
}
