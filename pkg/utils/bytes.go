package utils

import "unsafe"

// UnsafeToString []byte to string 无需额外空间
func UnsafeToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func SafeToString(b []byte) string {
	return string(b)
}
