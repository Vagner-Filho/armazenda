package report_router

import (
	"armazenda/service/user_service"
	report_view "armazenda/view/report"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getRelatorioPage(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	c.HTML(http.StatusOK, "relatorio.html", report_view.GetReportPage(farm))
}

func UseReportRoutes(router *gin.Engine) {
	router.GET("/relatorio", getRelatorioPage)
}
