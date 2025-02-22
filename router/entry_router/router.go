package entry_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/service/entry_service"
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

func GetEntryForm(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}

	entryForm, toasts := entry_view.GetExistingEntryForm(uint32(converted))

	for _, t := range toasts {
		if t != nil {
			c.Header("HX-Trigger", string(t.ToJson()))
		}
	}

	c.HTML(
		http.StatusOK,
		"entry-form",
		entryForm,
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

	toast := entry_service.DeleteEntry(uint32(converted))
	c.Header("HX-Trigger", string(toast.ToJson()))
	c.Status(http.StatusOK)
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

	//rawEntries, err := entry_model.FilterEntries(entryFilter)

	//if err != nil {
	//	c.HTML(http.StatusBadRequest, "toast", err.Error())
	//	return
	//}

	//if len(rawEntries) == 0 {
	//	c.HTML(http.StatusOK, "no-entry-found-for-filter", gin.H{})
	//	return
	//}

	c.HTML(http.StatusOK, "entry-table", gin.H{})
}

func GetEntryFiltersForm(c *gin.Context) {
	c.HTML(http.StatusOK, "entry-filter-form", entry_view.GetFiltersForm())
	return
}

func GetEmptyEntryForm(c *gin.Context) {
	formMembers, toasts := entry_view.GetEntryForm()

	for _, t := range toasts {
		if t != nil {
			c.Header("HX-Trigger", string(t.ToJson()))
		}
	}
	c.HTML(http.StatusOK, "entry-form", formMembers)
}
