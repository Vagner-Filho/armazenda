package stats_router

import (
	"armazenda/view/stats"

	"github.com/gin-gonic/gin"
)

func UseStatsRoutes(router *gin.Engine) {
	statsGroup := router.Group("/stats")
	{
		statsGroup.GET("/top-supplier", stats.TopSupplierCard)
		statsGroup.GET("/top-buyer", stats.TopBuyerCard)
		statsGroup.GET("/most-frequent-supplier", stats.MostFrequentSupplierCard)
		statsGroup.GET("/best-quality-supplier", stats.BestQualitySupplierCard)
		statsGroup.GET("/worst-quality-supplier", stats.WorstQualitySupplierCard)
	}
	router.GET("/analise", stats.GetAnalysisPage)
	router.GET("/analise/most-productive-field", stats.GetProductiveFields)
}
