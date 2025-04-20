package utils

import (
	"armazenda/service/user_service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GetReadableDate(date int64) string {
	return time.UnixMilli(date).Format("02/Jan/2006 - 03:04")
}

const TimeLayout string = "2006-01-02T15:04"
const DBTimeWithoutTimeZone string = "2006-01-02 03:04:05"

func GetFarmFromToken(c *gin.Context) uint32 {
	str, _ := c.Cookie("session_id")

	token, _ := jwt.ParseWithClaims(str, &user_service.ArmazendaUserClaims{}, func(token *jwt.Token) (any, error) {
		return []byte("secret-key"), nil
	})

	claims := token.Claims.(*user_service.ArmazendaUserClaims)
	return claims.Farm
}
