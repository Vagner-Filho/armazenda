package utils

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func GetReadableDate(date int64) string {
	return time.UnixMilli(date).Format("02/Jan/2006 - 03:04")
}

const TimeLayout string = "2006-01-02T15:04"
const DBTimeWithoutTimeZone string = "2006-01-02 03:04:05"
const DBDateOnly string = "2006-01-02"

func ParseMetaFromForm(c *gin.Context) (map[string]interface{}, bool) {
	meta := make(map[string]any)

	empty := true
	for i := range 30 {
		metaKeyAccessor := fmt.Sprintf("meta[%d].key", i)
		metaValueAccessor := fmt.Sprintf("meta[%d].value", i)

		metaKey := c.PostForm(metaKeyAccessor)
		metaValue := c.PostForm(metaValueAccessor)

		if metaKey == "" && metaValue == "" {
			break
		}

		if metaKey == "" || metaValue == "" {
			continue
		}

		empty = false
		meta[metaKey] = metaValue
	}

	return meta, empty
}
