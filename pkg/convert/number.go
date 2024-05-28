package cvt

import (
	"encoding/json"
	"math"
	"strconv"
)

// ToNumber 将interface安全转换为number.
func ToNumber(v interface{}) float64 {
	if v != nil {
		switch a := v.(type) {
		case []interface{}:
			if len(a) > 0 {
				return ToNumber(a[0])
			}
		case []map[string]interface{}:
			if len(a) > 0 {
				return ToNumber(a[0])
			}
		case float64:
			return a
		case string:
			if a == "" {
				return 0
			}
			if b, err := strconv.ParseBool(a); err == nil {
				return map[bool]float64{true: 1, false: 0}[b]
			}
			if f, err := strconv.ParseFloat(a, 64); err == nil {
				return f
			}
		case bool:
			if a {
				return 1
			}
			return 0
		case uint:
			return float64(a)
		case uint8:
			return float64(a)
		case uint16:
			return float64(a)
		case uint32:
			return float64(a)
		case uint64:
			return float64(a)
		case int:
			return float64(a)
		case int8:
			return float64(a)
		case int16:
			return float64(a)
		case int32:
			return float64(a)
		case int64:
			return float64(a)
		case float32:
			return float64(a)
		case json.Number:
			r, _ := a.Float64()
			return r
		default:
			return 0
		}
	}
	return 0
}

// ToInt 将interface安全转换为int.
func ToInt(v interface{}) int {
	return int(ToNumber(v))
}

// ToInt64 将interface安全转换为int64.
func ToInt64(v interface{}) int64 {
	return int64(ToNumber(v))
}

func Round(f float64, n int) float64 {
	pow10N := math.Pow10(n)
	// TODO +0.5 是为了四舍五入，如果不希望这样去掉这个
	return math.Trunc((f+0.5/pow10N)*pow10N) / pow10N
}
