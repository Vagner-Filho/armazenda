package humidity_progression_router

import (
	"fmt"
	"net/http"
	"strconv"

	entity_public "armazenda/entity/public"
	"armazenda/model/humidity_progression_model"
	"armazenda/service/user_service"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func UseHumidityProgressionHtmlRoutes(router *gin.Engine) {
	hpm := humidity_progression_model.GetHumidityProgressionModel()

	// GET /farm/config/progressions - Return progression table fragment
	router.GET("/farm/config/progressions", func(c *gin.Context) {
		sid, _ := c.Cookie("session_id")
		farm := user_service.GetFarmFromToken(sid)
		if farm == 0 {
			c.Status(http.StatusUnauthorized)
			return
		}

		progressions, modelErr := hpm.ListProgressions(uint32(farm))
		if modelErr != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		c.HTML(http.StatusOK, "progression-table", gin.H{"Progressions": progressions})
	})

	// GET /progressao/form - Return empty create dialog
	router.GET("/progressao/form", func(c *gin.Context) {
		sid, _ := c.Cookie("session_id")
		farm := user_service.GetFarmFromToken(sid)
		if farm == 0 {
			c.Status(http.StatusUnauthorized)
			return
		}

		c.HTML(http.StatusOK, "progression-form", gin.H{})
	})

	// GET /progressao/form/:id - Return edit dialog with data pre-filled
	router.GET("/progressao/form/:id", func(c *gin.Context) {
		sid, _ := c.Cookie("session_id")
		farm := user_service.GetFarmFromToken(sid)
		if farm == 0 {
			c.Status(http.StatusUnauthorized)
			return
		}

		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		progression, modelErr := hpm.GetProgression(uint32(id))
		if modelErr != nil {
			c.Status(http.StatusNotFound)
			return
		}

		// Verify this farm owns this progression (or it's system default)
		if progression.FarmId != nil && *progression.FarmId != uint32(farm) {
			c.Status(http.StatusForbidden)
			return
		}

		c.HTML(http.StatusOK, "progression-form", progression)
	})

	// POST /progressao - Create from form-data, return new row fragment
	router.POST("/progressao", func(c *gin.Context) {
		sid, _ := c.Cookie("session_id")
		farm := user_service.GetFarmFromToken(sid)
		if farm == 0 {
			c.Status(http.StatusUnauthorized)
			return
		}

		name := c.PostForm("name")
		if name == "" {
			c.Header("HX-Trigger", `{"toast":{"Message":"Nome é obrigatório","Type":2}}`)
			c.Status(http.StatusBadRequest)
			return
		}

		tiers, tierErr := parseTiersFromForm(c)
		if tierErr != "" {
			c.Header("HX-Trigger", fmt.Sprintf(`{"toast":{"Message":"%s","Type":2}}`, tierErr))
			c.Status(http.StatusBadRequest)
			return
		}

		id, modelErr := hpm.AddProgression(name, uint32(farm), tiers)
		if modelErr != nil {
			status := http.StatusBadRequest
			if modelErr.IsServerErr {
				status = http.StatusInternalServerError
			}
			c.Header("HX-Trigger", fmt.Sprintf(`{"toast":{"Message":"%s","Type":2}}`, modelErr.Message))
			c.Status(status)
			return
		}

		// Fetch the created progression with tiers for rendering
		progression, getErr := hpm.GetProgression(id)
		if getErr != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		t := entity_public.GetSuccessToast("Progressão criada com sucesso", "")
		c.Header("HX-Trigger", t.ToJsonStr())
		c.HTML(http.StatusCreated, "progression-row", progression)
	})

	// PUT /progressao/:id - Update from form-data, return updated row
	router.PUT("/progressao/:id", func(c *gin.Context) {
		sid, _ := c.Cookie("session_id")
		farm := user_service.GetFarmFromToken(sid)
		if farm == 0 {
			c.Status(http.StatusUnauthorized)
			return
		}

		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		name := c.PostForm("name")
		if name == "" {
			c.Header("HX-Trigger", `{"toast":{"Message":"Nome é obrigatório","Type":2}}`)
			c.Status(http.StatusBadRequest)
			return
		}

		tiers, tierErr := parseTiersFromForm(c)
		if tierErr != "" {
			c.Header("HX-Trigger", fmt.Sprintf(`{"toast":{"Message":"%s","Type":2}}`, tierErr))
			c.Status(http.StatusBadRequest)
			return
		}

		modelErr := hpm.UpdateProgression(uint32(id), name, tiers)
		if modelErr != nil {
			status := http.StatusBadRequest
			if modelErr.IsServerErr {
				status = http.StatusInternalServerError
			}
			c.Header("HX-Trigger", fmt.Sprintf(`{"toast":{"Message":"%s","Type":2}}`, modelErr.Message))
			c.Status(status)
			return
		}

		// Fetch the updated progression for rendering
		progression, getErr := hpm.GetProgression(uint32(id))
		if getErr != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Header("HX-Trigger", `{"toast":{"Message":"Progressão atualizada com sucesso","Type":0}}`)
		c.HTML(http.StatusOK, "progression-row", progression)
	})

	// DELETE /progressao/:id - Soft-delete, HTMX removes row from DOM
	router.DELETE("/progressao/:id", func(c *gin.Context) {
		sid, _ := c.Cookie("session_id")
		farm := user_service.GetFarmFromToken(sid)
		if farm == 0 {
			c.Status(http.StatusUnauthorized)
			return
		}

		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		modelErr := hpm.DeleteProgression(uint32(id))
		if modelErr != nil {
			status := http.StatusBadRequest
			if modelErr.IsServerErr {
				status = http.StatusInternalServerError
			}
			c.Header("HX-Trigger", fmt.Sprintf(`{"toast":{"Message":"%s","Type":2}}`, modelErr.Message))
			c.Status(status)
			return
		}

		c.Header("HX-Trigger", `{"toast":{"Message":"Progressão excluída com sucesso","Type":0}}`)
		c.Status(http.StatusOK)
	})
}

// parseTiersFromForm extracts tier data from indexed form fields
// Expected format: tiers[0].thresholdHumidity, tiers[0].discountValue, tiers[1]....
func parseTiersFromForm(c *gin.Context) ([]entity_public.HumidityProgressionTier, string) {
	var tiers []entity_public.HumidityProgressionTier

	for i := 0; i < 30; i++ {
		thresholdKey := fmt.Sprintf("tiers[%d].thresholdHumidity", i)
		discountKey := fmt.Sprintf("tiers[%d].discountValue", i)

		thresholdStr := c.PostForm(thresholdKey)
		discountStr := c.PostForm(discountKey)

		if thresholdStr == "" && discountStr == "" {
			break
		}

		if thresholdStr == "" || discountStr == "" {
			return nil, "Todas as faixas devem ter umidade e desconto preenchidos"
		}

		threshold, err := strconv.ParseFloat(thresholdStr, 64)
		if err != nil {
			return nil, fmt.Sprintf("Valor de umidade inválido na faixa %d", i+1)
		}

		discount, err := strconv.ParseFloat(discountStr, 64)
		if err != nil {
			return nil, fmt.Sprintf("Valor de desconto inválido na faixa %d", i+1)
		}

		tiers = append(tiers, entity_public.HumidityProgressionTier{
			ThresholdHumidity: decimal.NewFromFloat(threshold),
			DiscountValue:     decimal.NewFromFloat(discount),
		})
	}

	if len(tiers) < 1 {
		return nil, "A progressão deve ter pelo menos 1 faixa"
	}
	if len(tiers) > 30 {
		return nil, "A progressão não pode ter mais de 30 faixas"
	}

	return tiers, ""
}
