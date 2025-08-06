package farm_config_router

import (
	"armazenda/entity/public"
	farm_config_service "armazenda/service/farm_config"
	"armazenda/service/user_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UseFarmConfigRouter(router *gin.Engine) {
	router.GET("/farm/config", func(c *gin.Context) {
		// For now, let's assume a fixed farm ID. This should be retrieved from the user's session in a real application.
		sid, _ := c.Cookie("session_id")
		farm := user_service.GetFarmFromToken(sid)
		config, err := farm_config_service.GetFarmConfig(farm)
		if err != nil || config == nil {
			// If no config is found, render the page with default values.
			c.HTML(http.StatusOK, "config.html", gin.H{
				"Farm": &entity_public.Farm{
					Id:               farm,
					HumidityDiscount: 1.15,
					Address:          &entity_public.Address{},
				},
			})
			return
		}
		c.HTML(http.StatusOK, "config.html", gin.H{"Farm": config})
	})

	router.POST("/farm/config", func(c *gin.Context) {
		var form entity_public.Farm
		if err := c.ShouldBind(&form); err != nil {
			c.HTML(http.StatusBadRequest, "config.html", gin.H{"error": err.Error(), "Farm": &form})
			return
		}

		sid, _ := c.Cookie("session_id")
		farmID := user_service.GetFarmFromToken(sid)
		form.Id = uint32(farmID)

		if err := farm_config_service.UpsertFarmConfig(&form); err != nil {
			c.HTML(http.StatusInternalServerError, "config.html", gin.H{"error": "Failed to save configuration", "Farm": &form})
			return
		}

		c.HTML(http.StatusOK, "config.html", gin.H{"Farm": &form, "success": "Configuração salva com sucesso!"})
	})
}
