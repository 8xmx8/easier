package cvt

import (
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/exp/constraints"
)

// ToMap 将interface转换为map.
// 如果参数是字符串,则尝试进行json反序列化.
// 如果参数是interface{}/map[string]interface{}数组,则将数组第一个元素进行转换.
// 如果参数是其他类型，则先调用json进行序列化，再把结果进行反序列化.
func ToMap(v interface{}) (map[string]interface{}, error) {
	if v != nil {
		switch m := v.(type) {
		case string:
			var a = make(map[string]interface{})
			if err := json.Unmarshal([]byte(m), &a); err != nil {
				return a, fmt.Errorf("call ToMap: convert string failed: %v", err)
			}
			return a, nil
		case map[string]interface{}:
			return m, nil
		case []interface{}:
			if len(m) > 0 {
				return ToMap(m[0])
			}
		case []map[string]interface{}:
			if len(m) > 0 {
				return ToMap(m[0])
			}
		default:
			data, err := json.Marshal(v)
			var r = make(map[string]interface{})
			if err == nil {
				err = json.Unmarshal(data, &r)
			}
			if err != nil {
				return r, fmt.Errorf("call ToMap: json convert failed:  %v", err)
			}
			return r, nil
		}
	}
	return make(map[string]interface{}), errors.New("call ToMap: your input is nil")
}

func ToMaps(v interface{}) ([]map[string]interface{}, error) {
	if v != nil {
		switch m := v.(type) {
		case []interface{}:
			var mapData = make([]map[string]interface{}, 0)
			for i := range m {
				mapData = append(mapData, m[i].(map[string]interface{}))
			}
			return mapData, nil
		case []map[string]interface{}:
			return m, nil
		}
	}
	return make([]map[string]interface{}, 0), errors.New("call ToMap: your input is nil")
}

type MapKey interface {
	constraints.Ordered | ~bool
}

// GetKeys 获取Map的key值列表
func GetKeys[T MapKey, V any](data map[T]V) []T {
	keys := make([]T, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	return keys
}

// GetValues 获取Map的key值列表
func GetValues[K MapKey, V any](data map[K]V) []V {
	values := make([]V, 0, len(data))
	for _, v := range data {
		values = append(values, v)
	}
	return values
}
