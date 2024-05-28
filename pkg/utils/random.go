package utils

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

func GenerateRandomTaskID(time time.Time) string {
	rand.Seed(time.UnixNano()) // nolint
	timeStr := time.Format("20060102150405")
	randomSuffix := rand.Intn(9000) + 1000
	taskID := fmt.Sprintf("%s%d", timeStr, randomSuffix)
	return taskID
}

func GenerateRandomTaskName(name string) string {
	return fmt.Sprintf("%s-%s", name, time.Now().Format("2006-01-02 15:04:05"))
}

func GenerateRandomKey(length int) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes) // nolint
	if err != nil {
		return "" // 返回错误
	}
	return hex.EncodeToString(bytes)
}
