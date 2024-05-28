package utils

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"

	mapset "github.com/deckarep/golang-set"
)

type IfTrue interface {
	~bool | []string | any
}

func If[T IfTrue](b bool, trueVal, falseVal T) T {
	if b {
		return trueVal
	}
	return falseVal
}

func SliceSepByStep[T interface{}](originalSlice []T, step int) [][]T {
	length := len(originalSlice) / step
	if len(originalSlice)%step != 0 {
		length++
	}

	slicedSlice := make([][]T, length)
	for i := 0; i < length; i++ {
		start := i * step
		end := start + step
		if end > len(originalSlice) {
			end = len(originalSlice)
		}
		slicedSlice[i] = originalSlice[start:end]
	}
	return slicedSlice
}

// MD5
func GenerateMD5(s []byte) [16]byte {
	return md5.Sum(s)
}
func GenerateMD5ToHex(s []byte) string {
	return fmt.Sprintf("%x", GenerateMD5(s))
}

// base64编码
func Base64Encode(str string) string {
	bytes := []byte(str)
	return base64.StdEncoding.EncodeToString(bytes)
}

// base64解码
func Base64Decode(b64Str string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(b64Str)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func ReadToBase64(reader io.Reader) (string, error) {
	data, err := ReadToBytes(reader)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func ReadToBytes(reader io.Reader) (data []byte, err error) {
	buf := make([]byte, 1024)
	data = make([]byte, 0, 4096)
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}
		data = append(data, buf[:n]...)
	}
	return
}

func IsItemInSlice(s []any, k any) bool {
	return mapset.NewSetFromSlice(s).Contains(k)
}
