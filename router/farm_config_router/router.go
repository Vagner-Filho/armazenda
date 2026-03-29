package farm_config_router

import (
	"armazenda/entity/public"
	"armazenda/model/humidity_progression_model"
	farm_config_service "armazenda/service/farm_config"
	"armazenda/service/user_service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getProgressions(farm uint32) []entity_public.HumidityProgression {
	hpm := humidity_progression_model.GetHumidityProgressionModel()
	progressions, err := hpm.ListProgressions(farm)
	if err != nil {
		return []entity_public.HumidityProgression{}
	}
	return progressions
}

func UseFarmConfigRouter(router *gin.Engine) {
	router.GET("/farm/config", func(c *gin.Context) {
		sid, _ := c.Cookie("session_id")
		farm := user_service.GetFarmFromToken(sid)
		progressions := getProgressions(farm)

		config, err := farm_config_service.GetFarmConfig(farm)
		if err != nil || config == nil {
			nonce, _ := c.Get("csp_nonce")
			fmt.Printf("\n%v\n", nonce)
			c.HTML(http.StatusOK, "config.html", gin.H{
				"Farm":         &entity_public.Farm{Id: farm, Address: entity_public.Address{}},
				"Progressions": progressions,
				"CSPNonce":     nonce.(string),
			})
			return
		}
		nonce, _ := c.Get("csp_nonce")
		fmt.Printf("\n%v\n", nonce)
		c.HTML(http.StatusOK, "config.html", gin.H{"Farm": config, "Progressions": progressions, "CSPNonce": nonce.(string)})
	})

	router.PUT("/farm/config", func(c *gin.Context) {
		var form entity_public.Farm
		if err := c.ShouldBind(&form); err != nil {
			c.HTML(http.StatusBadRequest, "config-form", gin.H{"error": err.Error(), "Farm": &form, "Progressions": getProgressions(form.Id)})
			return
		}

		sid, _ := c.Cookie("session_id")
		farmID := user_service.GetFarmFromToken(sid)
		form.Id = uint32(farmID)

		if err := farm_config_service.UpsertFarmConfig(&form); err != nil {
			c.HTML(http.StatusInternalServerError, "config-form", gin.H{"error": "Failed to save configuration", "Farm": &form, "Progressions": getProgressions(farmID)})
			return
		}

		c.HTML(http.StatusOK, "config-form", gin.H{"Farm": &form, "Progressions": getProgressions(farmID), "success": "Configuração salva com sucesso!"})
	})

	router.PATCH("/api/farm/humidity-progression/:id", func(c *gin.Context) {
		sid, _ := c.Cookie("session_id")
		farmID := user_service.GetFarmFromToken(sid)
		if farmID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		// Parse progression ID from path
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid progression ID"})
			return
		}
		progressionID := uint32(id)

		if err := farm_config_service.SetFarmHumidityProgression(uint32(farmID), &progressionID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return updated progression table HTML fragment
		hpm := humidity_progression_model.GetHumidityProgressionModel()
		progressions, modelErr := hpm.ListProgressions(uint32(farmID))
		if modelErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": modelErr.Message})
			return
		}

		// Get current farm config to know which progression is default
		config, configErr := farm_config_service.GetFarmConfig(uint32(farmID))
		if configErr != nil || config == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get farm config"})
			return
		}

		c.HTML(http.StatusOK, "progression-table", gin.H{
			"Progressions":             progressions,
			"CurrentFarmProgressionId": config.HumidityProgressionId,
		})
	})
}
