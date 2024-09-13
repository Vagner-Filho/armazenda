package vehicle_router

import (
	"armazenda/service/vehicle_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetPlates(c *gin.Context) {
   c.HTML(http.StatusOK, "plateSelector", vehicle_service.GetPlates()) 
}

func PostPlate(c *gin.Context) {
    // query
}
