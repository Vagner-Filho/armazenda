package entry_router

import (
	"armazenda/model/entry_model"
	"armazenda/model/vehicle_model"
	"armazenda/service/entry_service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EntryForm struct {
	Product      entry_model.Grain `form:"product" binding:"gte=0"`
	Field        string            `form:"field" binding:"required"`
	Harvest      string            `form:"harvest" binding:"required"`
	Vehicle      string            `form:"vehicle"`
	VehiclePlate string            `form:"vehiclePlate" binding:"required"`
	GrossWeight  float64           `form:"grossWeight" binding:"required"`
	Tare         float64           `form:"tare" binding:"required"`
	NetWeight    float64           `form:"netWeight"`
	Humidity     string            `form:"humidity" binding:"required"`
	ArrivalDate  int64             `form:"arrivalDate" binding:"required"`
}

func GetEntries(c *gin.Context) {
	entries := entry_service.GetAllEntrySimplified()
	c.HTML(http.StatusOK, "grao.html", gin.H{
		"Entries": entries,
	})
}

func GetEntryForm(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}
	c.HTML(http.StatusOK, "addEntryDialog", entry_service.GetEntry(uint32(converted)))
}

func AddEntry(c *gin.Context) {
	var newEntry EntryForm
	err := c.Bind(&newEntry)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}
	ge := entry_model.Entry{
		Product:      newEntry.Product,
		Field:        entry_model.Field{},
		Harvest:      newEntry.Harvest,
		Waybill:      0,
		Vehicle:      vehicle_model.Vehicle{},
		ArrivalDate:  newEntry.ArrivalDate,
		GrossWeight:  newEntry.GrossWeight,
		Tare:         newEntry.Tare,
		NetWeight:    newEntry.NetWeight,
		Humidity:     newEntry.Humidity,
	}
	entry := entry_service.AddEntry(ge)
	c.HTML(http.StatusCreated, "entry", entry_service.MakeSimplifiedEntry(entry))
}

func DeleteEntry(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}

	c.String(http.StatusOK, "", entry_service.DeleteEntry(uint32(converted)))
}

func PutEntry(c *gin.Context) {
	id := c.Param("id")
	converted, parseErr := strconv.ParseUint(id, 10, 32)
	if parseErr != nil {
		c.String(http.StatusBadRequest, "", parseErr.Error())
		return
	}

	var newEntry EntryForm
	err := c.Bind(&newEntry)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	ge := entry_model.Entry{
		Product:      newEntry.Product,
		Field:        entry_model.Field{},
		Harvest:      newEntry.Harvest,
		Vehicle:      vehicle_model.Vehicle{},
		ArrivalDate:  newEntry.ArrivalDate,
		GrossWeight:  newEntry.GrossWeight,
		Tare:         newEntry.Tare,
		Humidity:     newEntry.Humidity,
		NetWeight:    newEntry.NetWeight,
		Waybill:      uint32(converted),
	}

	var updatedEntry = entry_service.PutEntry(ge)
	if updatedEntry != nil {
		c.HTML(http.StatusOK, "entry", entry_service.MakeSimplifiedEntry(*updatedEntry))
		return
	}
	c.HTML(500, "toast", "failed")
}
