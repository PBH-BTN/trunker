package conv

import (
	"unsafe"
)

// UnsafeBytesToString returns the byte slice as a volatile string
// THIS IS EVIL CODE.
// YOU HAVE BEEN WARNED.
func UnsafeBytesToString(b []byte) string {
	// same as strings.Builder::String()
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// UnsafeStringToBytes returns the string as a byte slice
// THIS IS EVIL CODE.
// YOU HAVE BEEN WARNED.
func UnsafeStringToBytes(s string) (b []byte) {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
