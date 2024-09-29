package utils

import "time"

func GetReadableDate(date int64) string {
	return time.UnixMilli(date).Format("02/Jan/2006 - 03:04")
}
