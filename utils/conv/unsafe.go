package conv

import (
	"reflect"
	"unsafe"
)

// UnsafeBytesToString returns the byte slice as a volatile string
// THIS IS EVIL CODE.
// YOU HAVE BEEN WARNED.
func UnsafeBytesToString(b []byte) string {
	// same as strings.Builder::String()
	return *(*string)(unsafe.Pointer(&b))
}

// UnsafeStringToBytes returns the string as a byte slice
// THIS IS EVIL CODE.
// YOU HAVE BEEN WARNED.
func UnsafeStringToBytes(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return b
}
