package user_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/service/user_service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {
	router.POST("/login", func(c *gin.Context) {
		var signInUser entity_public.SignInUser
		err := c.Bind(&signInUser)
		if err != nil {
			c.String(http.StatusBadRequest, "", err.Error())
			return
		}

		token, toast := user_service.Login(signInUser.Email, signInUser.Passwd)

		if toast != nil {
			c.Header("HX-Trigger", string(toast.ToJson()))
		}

		if len(token) > 0 {
			fmt.Printf("\n%v\n", token)
			c.Header("HX-Redirect", "/romaneio")
			c.Status(http.StatusOK)
		}
	})
}

func GetUserForm(c *gin.Context) {
	c.HTML(http.StatusOK, "user-form", gin.H{})
	return
}
