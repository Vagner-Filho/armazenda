package analysis_router

import (
	"armazenda/service/analysis_service"
	"armazenda/service/user_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getAnalysisPage(c *gin.Context) {
	c.HTML(http.StatusOK, "analise.html", gin.H{})
}

func getProductiveFields(c *gin.Context) {
	sessionCookie, _ := c.Request.Cookie("session_id")
	farmId := user_service.GetFarmFromToken(sessionCookie.Value)
	fields, toast := analysis_service.GetProductiveFields(farmId)
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	c.HTML(http.StatusOK, "most-productive-field", fields)
}

func UseAnalysisRoutes(router *gin.Engine) {
	router.GET("/analise", getAnalysisPage)
	router.GET("/analise/most-productive-field", getProductiveFields)
}
