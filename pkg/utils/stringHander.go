package utils

import (
	"encoding/json"
	"regexp"
	"strings"
)

func GetSplit(input string) []string {
	re := regexp.MustCompile(`\n+`)
	input = re.ReplaceAllString(input, "\n")

	// 去除开头和结尾的\n
	input = strings.Trim(input, "\n")
	domains := strings.Split(input, "\n")
	return domains
}

func GetSplitD(input string) []string {
	re := regexp.MustCompile(`,+`)
	input = re.ReplaceAllString(input, ",")

	// 去除开头和结尾的\n
	input = strings.Trim(input, ",")
	domains := strings.Split(input, ",")
	return domains
}

func MakeJSONOutput(data interface{}) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return `{"error": "Failed to convert to JSON"}`
	}
	return string(jsonBytes)
}

// 判断一个string切片是否包含某个子字符串
func Contanis(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
