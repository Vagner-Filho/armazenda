package user_router

import (
	"armazenda/service/user_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {
	userService := &user_service.UserService{}

	router.POST("/validate", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		isValid := userService.ValidateCredentials(username, password)

		if isValid {
			c.Header("HX-Redirect", "/romaneio")
			c.Status(http.StatusOK)
		} else {
			c.HTML(http.StatusOK, "error-message.html", gin.H{})
		}
	})
}

func GetUserForm(c *gin.Context) {
	c.HTML(http.StatusOK, "user-form", gin.H{})
	return
}
