package stats

import (
	"armazenda/entity/public"
	"armazenda/service/stats_service"
	"armazenda/service/user_service"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FieldProductChartsViewData struct {
	NominalLabelsJSON    string
	NominalDatasetsJSON  string
	RelativeLabelsJSON   string
	RelativeDatasetsJSON string
	HasData              bool
}

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
		"TierKey":  user_service.GetTierKeyFromContext(c),
	})
}

// buildFieldProductSeries converts (field, product) rows into a grouped-bar
// chart payload: labels are fields (in query order) and one dataset per
// product, values aligned to labels with zero fill for absent pairs.
func buildFieldProductSeries(rows []entity_public.FieldProductTotal, metric func(entity_public.FieldProductTotal) float64) entity_public.FieldProductSeries {
	series := entity_public.FieldProductSeries{
		Labels:   []string{},
		Datasets: []entity_public.ProductDataset{},
	}
	labelIndex := make(map[string]int)
	datasetIndex := make(map[string]int)

	for _, row := range rows {
		li, ok := labelIndex[row.FieldName]
		if !ok {
			li = len(series.Labels)
			labelIndex[row.FieldName] = li
			series.Labels = append(series.Labels, row.FieldName)
			// extend every existing dataset with a zero for the new field
			for di := range series.Datasets {
				series.Datasets[di].Values = append(series.Datasets[di].Values, 0)
			}
		}

		di, ok := datasetIndex[row.ProductName]
		if !ok {
			di = len(series.Datasets)
			datasetIndex[row.ProductName] = di
			values := make([]float64, len(series.Labels))
			values[li] = metric(row)
			series.Datasets = append(series.Datasets, entity_public.ProductDataset{
				Product: row.ProductName,
				Values:  values,
			})
			continue
		}

		series.Datasets[di].Values[li] = metric(row)
	}

	return series
}

func GetFieldProductCharts(c *gin.Context) {
	sessionCookie, _ := c.Request.Cookie("session_id")
	farmId := user_service.GetFarmFromToken(sessionCookie.Value)
	charts, toast := stats_service.GetFieldProductTotals(farmId)
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	nominalSeries := buildFieldProductSeries(charts.Nominal, func(r entity_public.FieldProductTotal) float64 {
		return r.TotalWeight
	})
	relativeSeries := buildFieldProductSeries(charts.Relative, func(r entity_public.FieldProductTotal) float64 {
		return r.Productivity
	})

	nominalLabelsJSON, _ := json.Marshal(nominalSeries.Labels)
	nominalDatasetsJSON, _ := json.Marshal(nominalSeries.Datasets)
	relativeLabelsJSON, _ := json.Marshal(relativeSeries.Labels)
	relativeDatasetsJSON, _ := json.Marshal(relativeSeries.Datasets)

	viewData := FieldProductChartsViewData{
		NominalLabelsJSON:    string(nominalLabelsJSON),
		NominalDatasetsJSON:  string(nominalDatasetsJSON),
		RelativeLabelsJSON:   string(relativeLabelsJSON),
		RelativeDatasetsJSON: string(relativeDatasetsJSON),
		HasData:              len(charts.Nominal) > 0 || len(charts.Relative) > 0,
	}

	c.HTML(http.StatusOK, "field-product-charts", viewData)
}

func ProductTotals(c *gin.Context) {
	sessionCookie, _ := c.Request.Cookie("session_id")
	farmId := user_service.GetFarmFromToken(sessionCookie.Value)
	totals, toast := stats_service.GetProductTotals(farmId)
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}
	c.HTML(http.StatusOK, "product-totals", totals)
}
