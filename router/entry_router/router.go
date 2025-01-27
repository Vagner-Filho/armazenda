package entry_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/entry_model"
	"armazenda/model/field_model"
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

func GetRomaneioPage(c *gin.Context) {
	c.HTML(http.StatusOK, "romaneio.html", entry_view.GetEntryContent())
}

func GetEntryContent(c *gin.Context) {
	c.HTML(http.StatusOK, "entry-content", entry_view.GetEntryContent())
}

type PopulatedEntryForm struct {
	Entry    entity_public.Entry
	Fields   []entity_public.Field
	Vehicles []entity_public.Vehicle
}

func GetEntryForm(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}

	entry := entry_service.GetEntry(uint32(converted))

	fModel, _ := field_model.GetFieldModel()
	fields, _ := fModel.GetFields()
	for i, field := range fields {
		if field.Id == entry.Field {
			fields[i].Selected = true
		}
	}

	vehicles, _ := vehicle_service.GetVehicles()
	for i, vehicle := range vehicles {
		if entry.Vehicle == vehicle.Plate {
			vehicles[i].Selected = true
		}
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

	entry, toast := entry_service.AddEntry(newEntry)
	c.Header("HX-Trigger", string(toast.ToJson()))

	if toast.Type == entity_public.WarningToast {
		c.Status(http.StatusBadRequest)
		return
	}
	if toast.Type == entity_public.ErrorToast {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.HTML(http.StatusCreated, "entry-list-item", entry)
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
	//id := c.Param("id")
	//converted, parseErr := strconv.ParseUint(id, 10, 32)
	//if parseErr != nil {
	//	c.String(http.StatusBadRequest, "", parseErr.Error())
	//	return
	//}

	var entry entity_public.Entry
	err := c.Bind(&entry)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	var updatedEntry, toast = entry_service.PutEntry(entry)
	c.Header("HX-Trigger", string(toast.ToJson()))

	if toast.Type == entity_public.WarningToast {
		c.Status(http.StatusBadRequest)
		return
	}
	if toast.Type == entity_public.ErrorToast {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.HTML(http.StatusOK, "entry-list-item", updatedEntry)
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

	c.HTML(http.StatusOK, "entry-table", rawEntries)
}

func GetEntryFiltersForm(c *gin.Context) {
	c.HTML(http.StatusOK, "entry-filter-form", entry_view.GetFiltersForm())
	return
}
