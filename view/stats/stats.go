package stats

import (
	"armazenda/service/stats_service"
	"armazenda/service/user_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func TopSupplierCard(c *gin.Context) {
	sessionCookie, _ := c.Request.Cookie("session_id")
	farmId := user_service.GetFarmFromToken(sessionCookie.Value)
	stat, toast := stats_service.GetTopSupplierStat(farmId)
	if toast != nil {
		c.HTML(http.StatusOK, "toast", toast)
		return
	}
	c.HTML(http.StatusOK, "top-supplier-card.html", stat)
}

func TopBuyerCard(c *gin.Context) {
	sessionCookie, _ := c.Request.Cookie("session_id")
	farmId := user_service.GetFarmFromToken(sessionCookie.Value)
	stat, toast := stats_service.GetTopBuyerStat(farmId)
	if toast != nil {
		c.HTML(http.StatusOK, "toast", toast)
		return
	}
	c.HTML(http.StatusOK, "top-buyer-card.html", stat)
}

func MostFrequentSupplierCard(c *gin.Context) {
	sessionCookie, _ := c.Request.Cookie("session_id")
	farmId := user_service.GetFarmFromToken(sessionCookie.Value)
	stat, toast := stats_service.GetMostFrequentSupplierStat(farmId)
	if toast != nil {
		c.HTML(http.StatusOK, "toast", toast)
		return
	}
	c.HTML(http.StatusOK, "most-frequent-supplier-card.html", stat)
}

func BestQualitySupplierCard(c *gin.Context) {
	sessionCookie, _ := c.Request.Cookie("session_id")
	farmId := user_service.GetFarmFromToken(sessionCookie.Value)
	stat, toast := stats_service.GetBestQualitySupplierStat(farmId)
	if toast != nil {
		c.HTML(http.StatusOK, "toast", toast)
		return
	}
	c.HTML(http.StatusOK, "best-quality-supplier-card.html", stat)
}

func WorstQualitySupplierCard(c *gin.Context) {
	sessionCookie, _ := c.Request.Cookie("session_id")
	farmId := user_service.GetFarmFromToken(sessionCookie.Value)
	stat, toast := stats_service.GetWorstQualitySupplierStat(farmId)
	if toast != nil {
		c.HTML(http.StatusOK, "toast", toast)
		return
	}
	c.HTML(http.StatusOK, "worst-quality-supplier-card.html", stat)
}

func GetAnalysisPage(c *gin.Context) {
	nonce, exists := c.Get("csp_nonce")
	if exists == false {
		c.Status(http.StatusForbidden)
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}
	c.HTML(http.StatusOK, "analise.html", gin.H{
		"CSPNonce": nonce.(string),
	})
}

func GetProductiveFields(c *gin.Context) {
	sessionCookie, _ := c.Request.Cookie("session_id")
	farmId := user_service.GetFarmFromToken(sessionCookie.Value)
	fields, toast := stats_service.GetProductiveFields(farmId)
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	c.HTML(http.StatusOK, "most-productive-field", fields)
}
