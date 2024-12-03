package entry_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/entry_model"
	"armazenda/model/vehicle_model"
	"armazenda/service/entry_service"
	"armazenda/service/vehicle_service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EntryForm struct {
	Product     entity_public.Grain `form:"product" binding:"gte=0"`
	Field       uint32              `form:"field" binding:"required"`
	Harvest     string              `form:"harvest" binding:"required"`
	Vehicle     string              `form:"vehiclePlate"`
	GrossWeight float64             `form:"grossWeight" binding:"required"`
	Tare        float64             `form:"tare" binding:"required"`
	NetWeight   float64             `form:"netWeight"`
	Humidity    string              `form:"humidity" binding:"required"`
	ArrivalDate int64               `form:"arrivalDate" binding:"required"`
}

type FieldForm struct {
	Name string `form:"name" binding:"required"`
	Id   uint32 `form:"id"`
}

func GetEntries(c *gin.Context) {
	entries := entry_service.GetAllEntrySimplified()
	c.HTML(http.StatusOK, "romaneio.html", entries)
}

func GetEntriesTable(c *gin.Context) {
	c.HTML(http.StatusOK, "entry-table", entry_service.GetAllEntrySimplified())
}

type Field struct {
	Selected bool
	entry_model.Field
}

type Vehicle struct {
	Selected bool
	vehicle_model.Vehicle
}

type PopulatedEntryForm struct {
	Entry    entry_model.Entry
	Fields   []Field
	Vehicles []Vehicle
}

func GetEntryForm(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}

	entry := entry_service.GetEntry(uint32(converted))
	//fields := entry_service.GetFields()
	var fields []Field
	for _, field := range entry_service.GetFields() {
		newF := Field{}
		newF.Id = field.Id
		newF.Selected = field.Id == entry.Field
		newF.Name = field.Name
		fields = append(fields, newF)
	}
	//vehicles := vehicle_service.GetVehicles()

	var vehicles []Vehicle
	for _, vehicle := range vehicle_service.GetVehicles() {
		newV := Vehicle{}
		newV.Selected = entry.Vehicle == vehicle.Plate
		newV.Name = vehicle.Name
		newV.Plate = vehicle.Plate
		vehicles = append(vehicles, newV)
	}

	c.HTML(
		http.StatusOK,
		"add-entry-dialog",
		PopulatedEntryForm{Entry: entry, Fields: fields, Vehicles: vehicles},
	)
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
		Manifest:    0,
		Vehicle:     newEntry.Vehicle,
		ArrivalDate: newEntry.ArrivalDate,
		GrossWeight: newEntry.GrossWeight,
		Tare:        newEntry.Tare,
		NetWeight:   newEntry.NetWeight,
		Humidity:    newEntry.Humidity,
	}
	entry := entry_service.AddEntry(ge)
	c.HTML(http.StatusCreated, "entry-list-item", entry_service.MakeSimplifiedEntry(entry))
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
		Manifest:    uint32(converted),
	}

	var updatedEntry = entry_service.PutEntry(ge)
	if updatedEntry != nil {
		c.HTML(http.StatusOK, "entry-list-item", entry_service.MakeSimplifiedEntry(*updatedEntry))
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
	c.HTML(http.StatusCreated, "field-option", entry_model.Field{Name: newField.Name, Id: newId})
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

	c.HTML(http.StatusOK, "field-form", regexPattern)
}
