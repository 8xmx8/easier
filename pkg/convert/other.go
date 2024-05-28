package cvt

import (
	"reflect"
	"strconv"

	"github.com/spf13/cast"
)

// ToBoolean 将interface安全转换为bool.
func ToBoolean(v interface{}) bool {
	return cast.ToBool(v)
}

func IsNil(i interface{}) bool {
	vi := reflect.ValueOf(i)
	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}
	return false
}

func AtoiSlice(numbers []int) []string {
	strNumbers := make([]string, len(numbers))
	for i, num := range numbers {
		strNumbers[i] = strconv.Itoa(num)
	}
	return strNumbers
}
