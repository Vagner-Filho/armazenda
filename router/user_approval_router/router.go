package user_approval_router

import (
	"armazenda/service/user_approval_service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func UserApprovalRoutes(router *gin.Engine) {
	router.GET("/user/pending", func(c *gin.Context) {
		sessionCookie, err := c.Request.Cookie("session_id")
		if err != nil {
			c.HTML(http.StatusUnauthorized, "401", gin.H{})
			return
		}

		pendingUsers, err := user_approval_service.GetPendingUsers(sessionCookie.Value)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error-message", gin.H{"error": err.Error()})
			return
		}

		c.HTML(http.StatusOK, "pending-users", gin.H{"users": pendingUsers})
	})

	router.POST("/user/approve/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error-message", gin.H{"error": "Invalid user ID"})
			return
		}

		err = user_approval_service.ApproveUser(uint32(id))
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error-message", gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusOK)
	})

	router.POST("/user/decline/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error-message", gin.H{"error": "Invalid user ID"})
			return
		}

		err = user_approval_service.DeclineUser(uint32(id))
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error-message", gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusOK)
	})
}
