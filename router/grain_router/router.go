package grain_router

import (
	"armazenda/entity"
	"armazenda/service/grain_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EntryForm struct {
	Product      entity.Grain `form:"product" binding:"gte=0"`
	Field        string       `form:"field" binding:"required"`
	HarvestYear  int64        `form:"harvestYear" binding:"required"`
	Vehicle      string       `form:"vehicle" binding:"required"`
	VehiclePlate string       `form:"vehiclePlate" binding:"required"`
	GrossWeight  float64      `form:"grossWeight" binding:"required"`
	Tare         float64      `form:"tare" binding:"required"`
	Humidity     string       `form:"humidity" binding:"required"`
	ArrivalDate  int64        `form:"arrivalDate" binding:"required"`
}

func GetGrains(c *gin.Context) {
	grains := grain_service.GetAllGrainEntrySimplified()
	c.HTML(http.StatusOK, "grao.html", gin.H{
		"Entries": grains,
	})
}

func AddGrain(c *gin.Context) {
	var newEntry EntryForm
	err := c.Bind(&newEntry)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}
    entry := grain_service.AddGrainEntry(entity.GrainEntry(newEntry))
	c.HTML(http.StatusCreated, "entry", grain_service.MakeSimplifieidGrainEntry(entry))
}
