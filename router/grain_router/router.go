package grain_router

import (
	"armazenda/model/grain_model"
	"armazenda/service/grain_service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EntryForm struct {
	Product      grain_model.Grain `form:"product" binding:"gte=0"`
	Field        string            `form:"field" binding:"required"`
	Harvest      string            `form:"harvest" binding:"required"`
	Vehicle      string            `form:"vehicle"`
	VehiclePlate string            `form:"vehiclePlate" binding:"required"`
	GrossWeight  float64           `form:"grossWeight" binding:"required"`
	Tare         float64           `form:"tare" binding:"required"`
	Humidity     string            `form:"humidity" binding:"required"`
	ArrivalDate  int64             `form:"arrivalDate" binding:"required"`
}

func GetGrains(c *gin.Context) {
	grains := grain_service.GetAllGrainEntrySimplified()
	c.HTML(http.StatusOK, "grao.html", gin.H{
		"Entries": grains,
	})
}

func GetEntryForm(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}
	c.HTML(http.StatusOK, "addEntryDialog", grain_service.GetEntry(uint32(converted)))
}

func AddGrain(c *gin.Context) {
	var newEntry EntryForm
	err := c.Bind(&newEntry)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}
	ge := grain_model.GrainEntry{
		Product:      newEntry.Product,
		Field:        newEntry.Field,
		Harvest:      newEntry.Harvest,
		Waybill:      0,
		Vehicle:      newEntry.Vehicle,
		VehiclePlate: newEntry.VehiclePlate,
		ArrivalDate:  newEntry.ArrivalDate,
		GrossWeight:  newEntry.GrossWeight,
		Tare:         newEntry.Tare,
		Humidity:     newEntry.Humidity,
	}
	entry := grain_service.AddGrainEntry(ge)
	c.HTML(http.StatusCreated, "entry", grain_service.MakeSimplifiedGrainEntry(entry))
}

func DeleteGrain(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}

	c.String(http.StatusOK, "", grain_service.DeleteGrainEntry(uint32(converted)))
}
