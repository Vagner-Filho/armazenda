package user_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/service/user_service"
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

		credentials, toast := user_service.Login(signInUser.Cpf, signInUser.Passwd)

		if toast != nil {
			c.Header("HX-Trigger", string(toast.ToJson()))
		}

		if len(credentials.Token) > 0 {
			c.SetCookie("session_id", credentials.Token, 6000, "", "", true, true)
			c.SetCookie("username", credentials.Username, 6000, "", "", true, false)
			c.Header("HX-Redirect", "/romaneio")
			c.Status(http.StatusOK)
		}
	})

	router.POST("/user", func(c *gin.Context) {
		var newUser entity_public.NewUser
		err := c.Bind(&newUser)
		if err != nil {
			c.Status(http.StatusBadRequest)
			toast := entity_public.GetWarningToast("Preencha todos os campos", "")
			c.Header("HX-Trigger", string(toast.ToJson()))
			return
		}

		toast := user_service.Create(newUser)
		c.Header("HX-Trigger", string(toast.ToJson()))
		if toast.Type == 0 {
			c.Status(http.StatusCreated)
			c.Header("HX-Redirect", "/")
		} else if toast.Type == 1 {
			c.Status(http.StatusBadRequest)
		}
	})
	router.GET("/user/form", getUserForm)
}

func getUserForm(c *gin.Context) {
	c.HTML(http.StatusOK, "user-form", gin.H{})
	return
}
