package view

import (
	"armazenda/entity/public"
	"armazenda/service/farm_config"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ShowConfigPage(c *gin.Context, service *farm_config.Service) {
	// For now, let's assume a fixed farm ID. This should be retrieved from the user's session in a real application.
	farmID := uint32(1)
	config, err := service.GetFarmConfig(farmID)
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
}
