package entry_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/service/entry_service"
	"armazenda/service/user_service"
	entry_view "armazenda/view/entry"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FieldForm struct {
	Name string `form:"name" binding:"required"`
	Id   uint32 `form:"id"`
}

func getRomaneioPage(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	c.HTML(http.StatusOK, "romaneio.html", entry_view.GetEntryContent(farm))
}

func getEntryContent(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	c.HTML(http.StatusOK, "entry-content", entry_view.GetEntryContent(farm))
}

func getEntryForm(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	entryForm, toasts := entry_view.GetExistingEntryForm(uint32(converted), farm)
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

func addEntry(c *gin.Context) {
	var newEntry entity_public.Entry
	err := c.Bind(&newEntry)
	if err != nil {
		toast := entity_public.GetWarningToast(err.Error(), "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	newEntry.Farm = farm
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

func deleteEntry(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}

	toast := entry_service.DeleteEntry(uint32(converted))
	c.Header("HX-Trigger", string(toast.ToJson()))
	c.Status(http.StatusOK)
}

func putEntry(c *gin.Context) {
	id := c.Param("id")
	converted, parseErr := strconv.ParseUint(id, 10, 32)
	if parseErr != nil {
		c.String(http.StatusBadRequest, "", parseErr.Error())
		return
	}

	var entry entity_public.Entry
	err := c.Bind(&entry)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	entry.Id = uint32(converted)
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

func filterEntries(c *gin.Context) {
	var entryFilter entity_public.EntryFilter
	err := c.Bind(&entryFilter)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	entries, toast := entry_service.FilterEntries(entryFilter)
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
	}

	if len(entries) == 0 {
		c.HTML(http.StatusOK, "no-entry-found-for-filter", gin.H{})
		return
	}

	c.HTML(http.StatusOK, "entry-table", entries)
}

func getEntryFiltersForm(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	c.HTML(http.StatusOK, "entry-filter-form", entry_view.GetFiltersForm(farm))
	return
}

func getEmptyEntryForm(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	formMembers, toasts := entry_view.GetEntryForm(farm)

	for _, t := range toasts {
		if t != nil {
			c.Header("HX-Trigger", string(t.ToJson()))
		}
	}
	c.HTML(http.StatusOK, "entry-form", formMembers)
}

func UseEntryRoutes(router *gin.Engine) {
	router.GET("/romaneio", getRomaneioPage)
	router.GET("/entry/list", getEntryContent)
	router.GET("/entry/filters", getEntryFiltersForm)
	router.GET("/entry/form", getEmptyEntryForm)
	router.GET("/entry/form/:id", getEntryForm)
	router.POST("/entry", addEntry)
	router.PUT("/entry/:id", putEntry)
	router.DELETE("/entry/:id", deleteEntry)
	router.POST("/entry/filter", filterEntries)
}
