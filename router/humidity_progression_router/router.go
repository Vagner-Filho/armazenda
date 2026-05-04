package humidity_progression_router

import (
	"net/http"
	"strconv"

	entity_public "armazenda/entity/public"
	"armazenda/model/humidity_progression_model"
	"armazenda/service/user_service"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func UseHumidityProgressionRouter(router *gin.Engine) {
	humidityRouter := router.Group("/api/humidity-progressions")

	// List all progressions for the farm
	humidityRouter.GET("", func(c *gin.Context) {
		sid, err := c.Cookie("session_id")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		farm := user_service.GetFarmFromToken(sid)
		if farm == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		hpm := humidity_progression_model.GetHumidityProgressionModel()
		progressions, modelErr := hpm.ListProgressions(uint32(farm))
		if modelErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": modelErr.Message})
			return
		}

		c.JSON(http.StatusOK, progressions)
	})

	// Get a specific progression
	humidityRouter.GET("/:id", func(c *gin.Context) {
		sid, err := c.Cookie("session_id")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		farm := user_service.GetFarmFromToken(sid)
		if farm == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		hpm := humidity_progression_model.GetHumidityProgressionModel()
		progression, modelErr := hpm.GetProgression(uint32(id))
		if modelErr != nil {
			if modelErr.IsServerErr {
				c.JSON(http.StatusInternalServerError, gin.H{"error": modelErr.Message})
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": modelErr.Message})
			}
			return
		}

		c.JSON(http.StatusOK, progression)
	})

	// Create a new progression
	humidityRouter.POST("", func(c *gin.Context) {
		sid, err := c.Cookie("session_id")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		farm := user_service.GetFarmFromToken(sid)
		if farm == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		var request CreateProgressionRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate tier count
		if len(request.Tiers) < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A progressão deve ter pelo menos 1 faixa"})
			return
		}
		if len(request.Tiers) > 30 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A progressão não pode ter mais de 30 faixas"})
			return
		}

		// Convert tiers
		tiers := make([]entity_public.HumidityProgressionTier, len(request.Tiers))
		for i, tier := range request.Tiers {
			tiers[i] = entity_public.HumidityProgressionTier{
				ThresholdHumidity: decimal.NewFromFloat(tier.ThresholdHumidity),
				DiscountValue:     decimal.NewFromFloat(tier.DiscountValue),
			}
		}

		hpm := humidity_progression_model.GetHumidityProgressionModel()
		id, modelErr := hpm.AddProgression(request.Name, uint32(farm), tiers)
		if modelErr != nil {
			if modelErr.IsServerErr {
				c.JSON(http.StatusInternalServerError, gin.H{"error": modelErr.Message})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": modelErr.Message})
			}
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": id})
	})

	// Update a progression
	humidityRouter.PUT("/:id", func(c *gin.Context) {
		sid, err := c.Cookie("session_id")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		farm := user_service.GetFarmFromToken(sid)
		if farm == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var request UpdateProgressionRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate tier count
		if len(request.Tiers) < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A progressão deve ter pelo menos 1 faixa"})
			return
		}
		if len(request.Tiers) > 30 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A progressão não pode ter mais de 30 faixas"})
			return
		}

		// Convert tiers
		tiers := make([]entity_public.HumidityProgressionTier, len(request.Tiers))
		for i, tier := range request.Tiers {
			tiers[i] = entity_public.HumidityProgressionTier{
				ThresholdHumidity: decimal.NewFromFloat(tier.ThresholdHumidity),
				DiscountValue:     decimal.NewFromFloat(tier.DiscountValue),
			}
		}

		hpm := humidity_progression_model.GetHumidityProgressionModel()
		modelErr := hpm.UpdateProgression(uint32(id), request.Name, tiers)
		if modelErr != nil {
			if modelErr.IsServerErr {
				c.JSON(http.StatusInternalServerError, gin.H{"error": modelErr.Message})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": modelErr.Message})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Progressão atualizada com sucesso"})
	})

	// Delete a progression
	humidityRouter.DELETE("/:id", func(c *gin.Context) {
		sid, err := c.Cookie("session_id")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		farm := user_service.GetFarmFromToken(sid)
		if farm == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		hpm := humidity_progression_model.GetHumidityProgressionModel()
		modelErr := hpm.DeleteProgression(uint32(id))
		if modelErr != nil {
			if modelErr.IsServerErr {
				c.JSON(http.StatusInternalServerError, gin.H{"error": modelErr.Message})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": modelErr.Message})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Progressão excluída com sucesso"})
	})
}

// Request types
type CreateProgressionRequest struct {
	Name  string `json:"name" binding:"required"`
	Tiers []struct {
		ThresholdHumidity float64 `json:"thresholdHumidity" binding:"required"`
		DiscountValue     float64 `json:"discountValue" binding:"required"`
	} `json:"tiers" binding:"required,dive"`
}

type UpdateProgressionRequest struct {
	Name  string `json:"name" binding:"required"`
	Tiers []struct {
		ThresholdHumidity float64 `json:"thresholdHumidity" binding:"required"`
		DiscountValue     float64 `json:"discountValue" binding:"required"`
	} `json:"tiers" binding:"required,dive"`
}
