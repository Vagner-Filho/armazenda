package utils

import "time"

func GetReadableDate(date int64) string {
	return time.UnixMilli(date).Format("02/Jan/2006 - 03:04")
}

const TimeLayout string = "2006-01-02T15:04"
