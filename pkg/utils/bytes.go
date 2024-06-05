package utils

import (
	"github.com/vmihailenco/msgpack"
	"unsafe"
)

// UnsafeToString []byte to string 无需额外空间
func UnsafeToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func SafeToString(b []byte) string {
	return string(b)
}

// Marshal 更加小的序列化
func Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(&v)
}
func Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, &v)
}
