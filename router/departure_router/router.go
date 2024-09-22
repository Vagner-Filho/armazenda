package departure_router

import (
	"armazenda/service/departure_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetDepartures(c *gin.Context) {
    c.HTML(http.StatusOK, "departure-table", departure_service.GetDepartures())
}

func GetDepartureForm(c *gin.Context) {
    c.HTML(http.StatusOK, "departure-form", gin.H{})
}
