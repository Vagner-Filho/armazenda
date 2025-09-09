package report_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/service/user_service"
	report_view "armazenda/view/report"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getRelatorioPage(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)

	pageData := report_view.GetReportPage(farm)

	c.HTML(http.StatusOK, "relatorio.html", pageData)
}

func filterReport(c *gin.Context) {
	var reportFilter entity_public.ReportFilter
	err := c.Bind(&reportFilter)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	report, toast := report_view.FilterReport(reportFilter, farm)
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}
	c.HTML(http.StatusOK, "report-content", report)
}

func getFullReport(c *gin.Context) {
	var reportFilter entity_public.ReportFilter
	err := c.Bind(&reportFilter)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	report, toast := report_view.GetFullReport(reportFilter, farm)
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}
	c.HTML(http.StatusOK, "full-report.html", report)
}

func UseReportRoutes(router *gin.Engine) {
	router.GET("/relatorio", getRelatorioPage)
	router.POST("/report/filter", filterReport)
	router.POST("/report", getFullReport)
}
