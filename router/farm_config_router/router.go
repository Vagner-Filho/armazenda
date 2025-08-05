package farm_config_router

import (
	"armazenda/entity/public"
	"armazenda/service/farm_config"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func UseFarmConfigRouter(router *gin.Engine) {
	router.GET("/config", func(c *gin.Context) {
		// For now, let's assume a fixed farm ID. This should be retrieved from the user's session in a real application.
		farmID := uint32(1)
		config, err := farm_config_service.GetFarmConfig(farmID)
		if err != nil {
			// If no config is found, render the page with default values.
			c.HTML(http.StatusOK, "config.html", gin.H{
				"Farm": &entity_public.Farm{
					Id:               farmID,
					HumidityDiscount: 1.7,
					Address:          &entity_public.Address{},
				},
			})
			return
		}
		c.HTML(http.StatusOK, "config.html", gin.H{"Farm": config})
	})

	router.POST("/config", func(c *gin.Context) {
		var form entity_public.Farm
		if err := c.ShouldBind(&form); err != nil {
			c.HTML(http.StatusBadRequest, "config.html", gin.H{"error": err.Error(), "Farm": &form})
			return
		}

		// For now, let's assume a fixed farm ID. This should be retrieved from the user's session in a real application.
		farmIDStr, _ := c.GetPostForm("id")
		farmID, _ := strconv.ParseUint(farmIDStr, 10, 32)
		form.Id = uint32(farmID)

		if err := farm_config_service.UpsertFarmConfig(&form); err != nil {
			c.HTML(http.StatusInternalServerError, "config.html", gin.H{"error": "Failed to save configuration", "Farm": &form})
			return
		}

		c.Redirect(http.StatusFound, "/config")
	})
}
