package entry_router

import (
	"armazenda/model/entry_model"
	"armazenda/model/vehicle_model"
	"armazenda/service/entry_service"
	"armazenda/service/vehicle_service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EntryForm struct {
	Product      entry_model.Grain `form:"product" binding:"gte=0"`
	Field        uint32            `form:"field" binding:"required"`
	Harvest      string            `form:"harvest" binding:"required"`
	Vehicle      string            `form:"vehiclePlate"`
	GrossWeight  float64           `form:"grossWeight" binding:"required"`
	Tare         float64           `form:"tare" binding:"required"`
	NetWeight    float64           `form:"netWeight"`
	Humidity     string            `form:"humidity" binding:"required"`
	ArrivalDate  int64             `form:"arrivalDate" binding:"required"`
}

type FieldForm struct {
	Name string `form:"name" binding:"required"`
	Id   uint32 `form:"id"`
}

func GetEntries(c *gin.Context) {
	entries := entry_service.GetAllEntrySimplified()
	c.HTML(http.StatusOK, "grao.html", gin.H{
		"Entries": entries,
	})
}

type PopulatedEntryForm struct {
	Entry    entry_model.Entry
	Fields   []entry_model.Field
	Vehicles []vehicle_model.Vehicle
}

func GetEntryForm(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}

	entry := entry_service.GetEntry(uint32(converted))
	fields := entry_service.GetFields()
	vehicles := vehicle_service.GetVehicles()

	c.HTML(http.StatusOK, "addEntryDialog", PopulatedEntryForm{Entry: entry, Fields: fields, Vehicles: vehicles})
}

func AddEntry(c *gin.Context) {
	var newEntry EntryForm
	err := c.Bind(&newEntry)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}
	ge := entry_model.Entry{
		Product:     newEntry.Product,
		Field:       newEntry.Field,
		Harvest:     newEntry.Harvest,
		Waybill:     0,
		Vehicle:     newEntry.Vehicle,
		ArrivalDate: newEntry.ArrivalDate,
		GrossWeight: newEntry.GrossWeight,
		Tare:        newEntry.Tare,
		NetWeight:   newEntry.NetWeight,
		Humidity:    newEntry.Humidity,
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
		Product:     newEntry.Product,
		Field:       newEntry.Field,
		Harvest:     newEntry.Harvest,
		Vehicle:     newEntry.Vehicle,
		ArrivalDate: newEntry.ArrivalDate,
		GrossWeight: newEntry.GrossWeight,
		Tare:        newEntry.Tare,
		Humidity:    newEntry.Humidity,
		NetWeight:   newEntry.NetWeight,
		Waybill:     uint32(converted),
	}

	var updatedEntry = entry_service.PutEntry(ge)
	if updatedEntry != nil {
		c.HTML(http.StatusOK, "entry", entry_service.MakeSimplifiedEntry(*updatedEntry))
		return
	}
	c.HTML(500, "toast", "failed")
}

func GetFields() []entry_model.Field {
	return entry_service.GetFields()
}

func AddField(c *gin.Context) {
	var newField FieldForm
	err := c.Bind(&newField)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}
	if len(newField.Name) == 0 {
		c.HTML(http.StatusBadRequest, "", "")
		return
	}

	newId := entry_service.AddField(newField.Name)
	c.HTML(http.StatusCreated, "fieldOption", entry_model.Field{Name: newField.Name, Id: newId})
	return
}

func GetFieldForm(c *gin.Context) {
	var fields = entry_service.GetFields()
	var regexPattern string = "^(?!"
	for i, field := range fields {
		regexPattern += field.Name + "$"
		if i < len(fields)-1 {
			regexPattern += "|"
		}
	}
	regexPattern += ").*"

	c.HTML(http.StatusOK, "fieldForm", regexPattern)
}
