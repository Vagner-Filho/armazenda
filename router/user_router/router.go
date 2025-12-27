package user_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/user_model"
	"armazenda/service/user_service"
	"net/http"
	"strconv"

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
		switch toast.Type {
		case 0:
			c.Status(http.StatusCreated)
			c.Header("HX-Redirect", "/")
		case 1:
			c.Status(http.StatusBadRequest)
		}
	})
	router.GET("/user/form", getUserForm)
	router.GET("/users", getUsers)
	router.POST("/user/:id/make-admin", makeAdmin)
	router.POST("/user/:id/remove-admin", removeAdmin)
	router.DELETE("/user/:id", removeUser)
}

func getUserForm(c *gin.Context) {
	c.HTML(http.StatusOK, "user-form", gin.H{})
}

func getUsers(c *gin.Context) {
	sessionCookie, cookieErr := c.Request.Cookie("session_id")
	if cookieErr != nil {
		c.HTML(http.StatusUnauthorized, "401", gin.H{})
		return
	}

	verifyErr := user_service.VerifyToken(sessionCookie.Value)
	if verifyErr != nil {
		c.HTML(http.StatusUnauthorized, "401", gin.H{})
		return
	}

	farmId := user_service.GetFarmFromToken(sessionCookie.Value)
	users, err := user_service.GetUsersByFarm(farmId)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error-message", gin.H{"message": err.Error()})
		return
	}

	isAdmin := user_service.IsAdmin(sessionCookie.Value)
	c.HTML(http.StatusOK, "user-list", gin.H{"users": users, "isAdmin": isAdmin})
}

func makeAdmin(c *gin.Context) {
	sessionCookie, cookieErr := c.Request.Cookie("session_id")
	if cookieErr != nil {
		c.HTML(http.StatusUnauthorized, "401", gin.H{})
		return
	}

	verifyErr := user_service.VerifyToken(sessionCookie.Value)
	if verifyErr != nil {
		c.HTML(http.StatusUnauthorized, "401", gin.H{})
		return
	}

	if !user_service.IsAdmin(sessionCookie.Value) {
		c.String(http.StatusForbidden, "Acesso negado. Apenas administradores podem realizar esta ação.")
		return
	}

	userIdStr := c.Param("id")
	userId, parseErr := strconv.ParseUint(userIdStr, 10, 32)
	if parseErr != nil {
		c.String(http.StatusBadRequest, "ID de usuário inválido")
		return
	}

	err := user_service.MakeAdmin(uint32(userId))
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao promover usuário", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	toast := entity_public.GetSuccessToast("Usuário promovido a administrador", "")
	c.Header("HX-Trigger", string(toast.ToJson()))
}

func removeAdmin(c *gin.Context) {
	sessionCookie, cookieErr := c.Request.Cookie("session_id")
	if cookieErr != nil {
		c.HTML(http.StatusUnauthorized, "401", gin.H{})
		return
	}

	verifyErr := user_service.VerifyToken(sessionCookie.Value)
	if verifyErr != nil {
		c.HTML(http.StatusUnauthorized, "401", gin.H{})
		return
	}

	if !user_service.IsAdmin(sessionCookie.Value) {
		c.String(http.StatusForbidden, "Acesso negado. Apenas administradores podem realizar esta ação.")
		return
	}

	userIdStr := c.Param("id")
	userId, parseErr := strconv.ParseUint(userIdStr, 10, 32)
	if parseErr != nil {
		c.String(http.StatusBadRequest, "ID de usuário inválido")
		return
	}

	err := user_service.RemoveAdmin(uint32(userId))
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao remover administrador", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	toast := entity_public.GetSuccessToast("Privilégios de administrador removidos", "")
	c.Header("HX-Trigger", string(toast.ToJson()))
}

func removeUser(c *gin.Context) {
	sessionCookie, cookieErr := c.Request.Cookie("session_id")
	if cookieErr != nil {
		c.HTML(http.StatusUnauthorized, "401", gin.H{})
		return
	}

	verifyErr := user_service.VerifyToken(sessionCookie.Value)
	if verifyErr != nil {
		c.HTML(http.StatusUnauthorized, "401", gin.H{})
		return
	}

	if !user_service.IsAdmin(sessionCookie.Value) {
		c.String(http.StatusForbidden, "Acesso negado. Apenas administradores podem realizar esta ação.")
		return
	}

	userIdStr := c.Param("id")
	userId, parseErr := strconv.ParseUint(userIdStr, 10, 32)
	if parseErr != nil {
		c.String(http.StatusBadRequest, "ID de usuário inválido")
		return
	}

	um := user_model.GetUserModel()
	canRemove, err := um.CanRemoveAdmin(uint32(userId))
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao verificar usuário", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	if !canRemove {
		toast := entity_public.GetErrorToast("Não é possível remover o último administrador", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	err = user_service.RemoveUser(uint32(userId))
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao remover usuário", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	toast := entity_public.GetSuccessToast("Acesso do usuário removido", "")
	c.Header("HX-Trigger", string(toast.ToJson()))
}
