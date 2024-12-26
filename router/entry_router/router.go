package entry_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/entry_model"
	"armazenda/model/vehicle_model"
	"armazenda/service/entry_service"
	"armazenda/service/vehicle_service"
	entry_view "armazenda/view/entry"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FieldForm struct {
	Name string `form:"name" binding:"required"`
	Id   uint32 `form:"id"`
}

type entryFilters struct {
	Fields   []entry_model.Field
	Vehicles []vehicle_model.Vehicle
}

func GetRomaneioPage(c *gin.Context) {
	entries := entry_view.GetAllEntrySimplified()
	vehicles := vehicle_service.GetVehicles()
	fields := GetFields()
	c.HTML(http.StatusOK, "romaneio.html", gin.H{
		"Entries": entries,
		"Filters": entryFilters{
			Vehicles: vehicles,
			Fields:   fields,
		},
	})
}

func GetEntriesTable(c *gin.Context) {
	c.HTML(http.StatusOK, "entry-table", entry_view.GetAllEntrySimplified())
}

type PopulatedEntryForm struct {
	Entry    entity_public.Entry
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
	//fields := entry_service.GetFields()
	var fields []entry_model.Field
	for _, field := range entry_service.GetFields() {
		newF := entry_model.Field{}
		newF.Id = field.Id
		newF.Selected = field.Id == entry.Field
		newF.Name = field.Name
		fields = append(fields, newF)
	}
	//vehicles := vehicle_service.GetVehicles()

	var vehicles []vehicle_model.Vehicle
	for _, vehicle := range vehicle_service.GetVehicles() {
		newV := vehicle_model.Vehicle{}
		newV.Selected = entry.Vehicle == vehicle.Plate
		newV.Name = vehicle.Name
		newV.Plate = vehicle.Plate
		vehicles = append(vehicles, newV)
	}

	c.HTML(
		http.StatusOK,
		"entry-form",
		PopulatedEntryForm{Entry: entry, Fields: fields, Vehicles: vehicles},
	)
}

func AddEntry(c *gin.Context) {
	var newEntry entity_public.Entry
	err := c.Bind(&newEntry)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}
	ge := entity_public.Entry{
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
	c.HTML(http.StatusCreated, "entry-list-item", entry_view.MakeSimplifiedEntry(entry))
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

	var newEntry entity_public.Entry
	err := c.Bind(&newEntry)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	ge := entity_public.Entry{
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
		c.HTML(http.StatusOK, "entry-list-item", entry_view.MakeSimplifiedEntry(*updatedEntry))
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

func FilterEntries(c *gin.Context) {
	var entryFilter entity_public.EntryFilter
	err := c.Bind(&entryFilter)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	rawEntries, err := entry_model.FilterEntries(entryFilter)

	if err != nil {
		c.HTML(http.StatusBadRequest, "toast", err.Error())
		return
	}

	if len(rawEntries) == 0 {
		c.HTML(http.StatusOK, "no-entry-found-for-filter", gin.H{})
		return
	}

	var simpleEntries []entry_view.SimplifiedEntry

	for _, entry := range rawEntries {
		simpleEntries = append(simpleEntries, entry_view.MakeSimplifiedEntry(entry))
	}

	c.HTML(http.StatusOK, "entry-table", simpleEntries)
}

func GetEntryFiltersForm(c *gin.Context) {
	vehicles := vehicle_service.GetVehicles()
	fields := GetFields()
	filters := entryFilters{
		Vehicles: vehicles,
		Fields:   fields,
	}
	c.HTML(http.StatusOK, "entry-filter-form", filters)
	return
}
