package user_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/user_model"
	"armazenda/service/user_service"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
			fmt.Printf("%v", credentials.Username)
			c.SetCookie("session_id", credentials.Token, 6000, "", "", true, true)
			c.SetCookie("username", credentials.Username, 6000, "", "", true, false)
			c.SetCookie("farmId", fmt.Sprintf("%v", credentials.Farm), 6000, "", "", true, false)
			c.Header("HX-Redirect", "/romaneio")
			c.Status(http.StatusOK)
		}
	})

	router.GET("/auth/google/login", googleLogin)
	router.GET("/auth/google/callback", googleCallback)
	router.GET("/user/google-register", googleRegisterForm)
	router.POST("/user/google-register", googleRegister)

	router.GET("/auth/microsoft/login", microsoftLogin)
	router.GET("/auth/microsoft/callback", microsoftCallback)
	router.GET("/user/microsoft-register", microsoftRegisterForm)
	router.POST("/user/microsoft-register", microsoftRegister)

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
		switch toast.Type {
		case 0:
			var msg string
			var title string
			if strings.Contains(toast.Message, "criado") {
				title = "Conta criada com sucesso!"
				msg = "Seu usuário foi criado e você já pode entrar no sistema"
			} else {
				title = "Aguardando aprovação"
				msg = "Seu usuário foi criado e agora precisa ser aprovado pelo administrador para entrar"
			}
			c.HTML(http.StatusOK, "user-success-modal", gin.H{
				"Body":  msg,
				"Title": title,
			})
		case 1:
			c.Header("HX-Trigger", string(toast.ToJson()))
			c.Status(http.StatusBadRequest)
		}
	})
	router.GET("/user/form", getUserForm)
	router.GET("/users", getUsers)
	router.POST("/user/:id/make-admin", makeAdmin)
	router.POST("/user/:id/remove-admin", removeAdmin)
	router.POST("/user/:id/activate", activateUser)
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
	isAdmin := user_service.IsAdmin(sessionCookie.Value)
	users, err := user_service.GetUsersByFarm(farmId, isAdmin)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error-message", gin.H{"message": err.Error()})
		return
	}

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

	userIdStr := c.Param("id")
	userId, parseErr := strconv.ParseUint(userIdStr, 10, 32)
	if parseErr != nil {
		c.String(http.StatusBadRequest, "ID de usuário inválido")
		return
	}

	claims := user_service.GetClaimsFromToken(sessionCookie.Value)
	if claims == nil {

	}
	usr, toast := user_service.RemoveAdmin(uint32(userId), claims.Id)
	if toast.Type == entity_public.SuccessToast {
		c.Header("HX-Trigger", string(toast.ToJson()))
		c.HTML(http.StatusOK, "user-list-item", gin.H{
			"isAdmin": true,
			"user":    usr,
		})
		return
	}
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

	isAdmin, err := um.IsAdmin(uint32(userId))

	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao verificar permissão do usuário", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	if isAdmin == true {
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

func activateUser(c *gin.Context) {
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

	err := user_service.ActivateUser(uint32(userId))
	if err != nil {
		toast := entity_public.GetErrorToast("Erro ao ativar usuário", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	toast := entity_public.GetSuccessToast("Usuário ativado com sucesso", "")
	c.Header("HX-Trigger", string(toast.ToJson()))

	farmId := user_service.GetFarmFromToken(sessionCookie.Value)
	isAdmin := user_service.IsAdmin(sessionCookie.Value)
	users, err := user_service.GetUsersByFarm(farmId, isAdmin)
	if err != nil {
		return
	}

	for _, user := range users {
		if user.Id == uint32(userId) {
			c.HTML(http.StatusOK, "user-list-item", gin.H{"user": user, "isAdmin": isAdmin})
			return
		}
	}
}

func googleLogin(c *gin.Context) {
	url := user_service.GetGoogleLoginURL()
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func googleCallback(c *gin.Context) {
	code := c.Query("code")
	credentials, toast := user_service.LoginWithGoogle(code)

	if toast != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{"error": toast.Message})
		return
	}

	if len(credentials.Token) > 0 {
		c.SetCookie("session_id", credentials.Token, 6000, "", "", true, true)
		c.SetCookie("username", credentials.Username, 6000, "", "", true, false)
		c.SetCookie("farmId", fmt.Sprintf("%v", credentials.Farm), 6000, "", "", true, false)
		c.Redirect(http.StatusSeeOther, "/romaneio")
		return
	}

	if credentials.IsNewUser {
		preRegToken, err := user_service.CreatePreRegistrationToken(credentials.Email, credentials.Name)
		if err != nil {
			c.HTML(http.StatusOK, "login.html", gin.H{"error": "Erro ao iniciar cadastro"})
			return
		}

		c.SetCookie("pre_reg_token", preRegToken, 1800, "", "", true, true)
		c.Redirect(http.StatusSeeOther, "/user/google-register")
		return
	}
}

func googleRegisterForm(c *gin.Context) {
	cookie, err := c.Cookie("pre_reg_token")
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	claims, err := user_service.VerifyPreRegistrationToken(cookie)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	c.HTML(http.StatusOK, "oauth-registration.html", gin.H{
		"Email":        claims.Email,
		"Name":         claims.Name,
		"ProviderName": "Google",
		"PostUrl":      "/user/google-register",
	})
}

func microsoftLogin(c *gin.Context) {
	url := user_service.GetMicrosoftLoginURL()
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func microsoftCallback(c *gin.Context) {
	code := c.Query("code")
	credentials, toast := user_service.LoginWithMicrosoft(code)

	if toast != nil {
		fmt.Printf("%v", toast.Message)
		c.HTML(http.StatusOK, "login.html", gin.H{"error": toast.Message})
		return
	}

	if len(credentials.Token) > 0 {
		c.SetCookie("session_id", credentials.Token, 6000, "", "", true, true)
		c.SetCookie("username", credentials.Username, 6000, "", "", true, false)
		c.SetCookie("farmId", fmt.Sprintf("%v", credentials.Farm), 6000, "", "", true, false)
		c.Redirect(http.StatusSeeOther, "/romaneio")
		return
	}

	if credentials.IsNewUser {
		preRegToken, err := user_service.CreatePreRegistrationToken(credentials.Email, credentials.Name)
		if err != nil {
			c.HTML(http.StatusOK, "login.html", gin.H{"error": "Erro ao iniciar cadastro"})
			return
		}

		c.SetCookie("pre_reg_token", preRegToken, 1800, "", "", true, true)
		c.Redirect(http.StatusSeeOther, "/user/microsoft-register")
		return
	}
}

func microsoftRegisterForm(c *gin.Context) {
	cookie, err := c.Cookie("pre_reg_token")
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	claims, err := user_service.VerifyPreRegistrationToken(cookie)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	c.HTML(http.StatusOK, "oauth-registration.html", gin.H{
		"Email":        claims.Email,
		"Name":         claims.Name,
		"ProviderName": "Microsoft",
		"PostUrl":      "/user/microsoft-register",
	})
}

func microsoftRegister(c *gin.Context) {
	cookie, err := c.Cookie("pre_reg_token")
	if err != nil {
		c.Status(http.StatusUnauthorized)
		toast := entity_public.GetErrorToast("Sessão de cadastro expirada", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	claims, err := user_service.VerifyPreRegistrationToken(cookie)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		toast := entity_public.GetErrorToast("Sessão de cadastro inválida", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	var form struct {
		Cpf               string `form:"cpf" binding:"len=11"`
		InscricaoEstadual string `form:"inscricaoEstadual" binding:"required"`
		Passwd            string `form:"passwd" binding:"required"`
		PasswdConfirm     string `form:"passwdConfirm" binding:"required"`
		Role              string `form:"role" binding:"oneof=admin user"`
	}

	err = c.Bind(&form)
	if err != nil {
		c.Status(http.StatusBadRequest)
		toast := entity_public.GetWarningToast("Preencha todos os campos", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	newUser := entity_public.NewUser{
		User: entity_public.User{
			Email:             claims.Email,
			Name:              claims.Name,
			Cpf:               form.Cpf,
			InscricaoEstadual: form.InscricaoEstadual,
			Passwd:            form.Passwd,
			Role:              form.Role,
		},
		PasswdConfirm: form.PasswdConfirm,
	}

	toast := user_service.Create(newUser)

	if toast.Type == 0 { // Success
		c.SetCookie("pre_reg_token", "", -1, "", "", true, true)
		c.Header("HX-Redirect", "/")
	}

	c.Header("HX-Trigger", string(toast.ToJson()))
}

func googleRegister(c *gin.Context) {
	cookie, err := c.Cookie("pre_reg_token")
	if err != nil {
		c.Status(http.StatusUnauthorized)
		toast := entity_public.GetErrorToast("Sessão de cadastro expirada", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	claims, err := user_service.VerifyPreRegistrationToken(cookie)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		toast := entity_public.GetErrorToast("Sessão de cadastro inválida", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	// Use a partial struct for binding to avoid 'required' validation on Email/Name
	var form struct {
		Cpf               string `form:"cpf" binding:"len=11"`
		InscricaoEstadual string `form:"inscricaoEstadual" binding:"required"`
		Passwd            string `form:"passwd" binding:"required"`
		PasswdConfirm     string `form:"passwdConfirm" binding:"required"`
		Role              string `form:"role" binding:"oneof=admin user"`
	}

	err = c.Bind(&form)
	if err != nil {
		c.Status(http.StatusBadRequest)
		toast := entity_public.GetWarningToast("Preencha todos os campos", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	// Construct the full NewUser object
	newUser := entity_public.NewUser{
		User: entity_public.User{
			Email:             claims.Email,
			Name:              claims.Name,
			Cpf:               form.Cpf,
			InscricaoEstadual: form.InscricaoEstadual,
			Passwd:            form.Passwd,
			Role:              form.Role,
		},
		PasswdConfirm: form.PasswdConfirm,
	}

	toast := user_service.Create(newUser)

	if toast.Type == 0 { // Success
		// Clear pre-reg token
		c.SetCookie("pre_reg_token", "", -1, "", "", true, true)

		c.Header("HX-Redirect", "/")

		// Send a toast to appear on the login page (via session/cookie or just rely on them logging in)
		// Since we redirect, HTMX won't show the toast on the current page.
		// But the user will see the login page.
		// We can add a query param to login page to show success message if we wanted,
		// but standard toast flow might be tricky with full redirect.
		// However, user_service.Create might return "User sent for approval".

		// Let's pass the toast via header so if it was an HTMX request that followed redirect it might show?
		// Actually hx-boost isn't used there.

		// For now, let's just trigger the toast. HTMX will handle the redirect.
	}

	c.Header("HX-Trigger", string(toast.ToJson()))
}
