package utils

import (
	"time"
)

const (
	LogTimeFormatMill        = "2006-01-02 15:04:05.9999"
	LogTimeFormatSecond      = "2006-01-02 15:04:05"
	LogTimeFormat            = "2006-01-02 15:04"
	StrTimeFormat            = "20060102150405"
	StrTimeFormatMill        = "200601021504059999"
	YYYYMMDDHHmmSSTimeFormat = "2006年01月02日 15:04:05"
	YYYYMMDDTimeDotFormat    = "2006.01.02"
	MonthFormat              = "2006-01"
	ZreoS                    = "0s"
	YYYYMMDDFormat           = "2006-01-02"
	EarlyTime                = 1019101055000
	LateTime                 = 4111702655000
)

// 解析时间, 如果解析错误, 返回"当前时间"
func ParseTime(layout, timeStr string) (t time.Time) {
	var err error
	if t, err = time.Parse(layout, timeStr); err != nil {
		t = time.Now()
	}
	return
}

// MillisToTime 毫秒数据戳转time
func MillisToTime(millis int64) time.Time {
	nanos := millis * int64(time.Millisecond)
	return time.Unix(0, nanos)
}

// GetDayInterval 获取当前毫秒时间戳所在天的开始和结束shike
func GetDayInterval(millis int64) (startOfDay, endOfDay time.Time) {
	timestamp := time.Unix(millis/1000, (millis%1000)*1e6)
	// 获取当天的开始时间（午夜）
	startOfDay = time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), 0, 0, 0, 0, timestamp.Location())
	// 获取当天的结束时间（次日的午夜前一秒）
	endOfDay = startOfDay.AddDate(0, 0, 1).Add(-1 * time.Second)
	return
}

// 根据时间获得毫秒级时间戳
func GetTimeMillis(tstr string) int64 {
	t := ParseTime(LogTimeFormat, tstr)
	return t.UnixNano() / int64(time.Millisecond)
}
