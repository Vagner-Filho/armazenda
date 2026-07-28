package entry_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/crop_model"
	"armazenda/model/entry_model"
	"armazenda/model/farm_config_model"
	"armazenda/model/humidity_progression_model"
	"armazenda/model/person_model"
	"armazenda/model/product_model"
	"armazenda/service/entry_service"
	"armazenda/service/user_service"
	entry_view "armazenda/view/entry"
	"fmt"
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
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}

	nonce, exists := c.Get("csp_nonce")
	if exists == false {
		c.Status(http.StatusForbidden)
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	content := entry_view.GetEntryContent(farm, page)
	content.CSPNonce = nonce.(string)
	content.TierKey = user_service.GetTierKeyFromContext(c)
	c.HTML(http.StatusOK, "romaneio.html", content)
}

func getEntryContent(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}

	nonce, exists := c.Get("csp_nonce")
	if exists == false {
		c.Status(http.StatusForbidden)
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	content := entry_view.GetEntryContent(farm, page)
	content.CSPNonce = nonce.(string)
	c.HTML(http.StatusOK, "entry-content", content)
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

	nonce, _ := c.Get("csp_nonce")
	entryForm.CSPNonce = nonce.(string)
	c.HTML(
		http.StatusOK,
		"entry-form",
		entryForm,
	)
}

func addEntryDraft(c *gin.Context) {
	var newEntry entity_public.EntryDraft
	err := c.Bind(&newEntry)
	if err != nil {
		toast := entity_public.GetWarningToast(err.Error(), "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	newEntry.Farm = farm
	entry, toast := entry_service.AddEntryDraft(newEntry)
	c.Header("HX-Trigger", string(toast.ToJson()))

	if toast.Type == entity_public.WarningToast {
		c.Header("HX-Trigger", string(toast.ToJson()))
		c.Status(http.StatusBadRequest)
		return
	}
	if toast.Type == entity_public.ErrorToast {
		c.Header("HX-Trigger", string(toast.ToJson()))
		c.Status(http.StatusInternalServerError)
		return
	}
	c.HTML(http.StatusCreated, "entry-draft-list-item", entry)
}

func addEntry(c *gin.Context) {
	var newEntry entity_public.Entry
	err := c.Bind(&newEntry)
	if err != nil {
		fmt.Print(err.Error())
		toast := entity_public.GetWarningToast(err.Error(), "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	newEntry.Farm = farm

	prod_m := product_model.GetProductModel()
	pm := person_model.GetPersonModel()
	cm := crop_model.GetCropModel()
	em := entry_model.GetEntryModel()
	hpm := humidity_progression_model.GetHumidityProgressionModel()
	fcm := farm_config_model.GetFarmConfigModel()
	entry, toast := entry_service.AddEntry(newEntry, em, pm, prod_m, cm, hpm, fcm)
	c.Header("HX-Trigger", toast.ToJsonStr())

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

func putEntryDraft(c *gin.Context) {
	id := c.Param("id")
	converted, parseErr := strconv.ParseUint(id, 10, 32)
	if parseErr != nil {
		c.String(http.StatusBadRequest, "", parseErr.Error())
		return
	}

	var entryDraft entity_public.EntryDraft
	err := c.Bind(&entryDraft)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	entryDraft.Id = uint32(converted)
	var updatedEntryDraft, toast = entry_service.PutEntryDraft(entryDraft)
	c.Header("HX-Trigger", string(toast.ToJson()))

	if toast.Type == entity_public.WarningToast {
		c.Status(http.StatusBadRequest)
		return
	}
	if toast.Type == entity_public.ErrorToast {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.HTML(http.StatusOK, "entry-draft-list-item", updatedEntryDraft)
}

func filterEntries(c *gin.Context) {
	var entryFilter entity_public.EntryFilter
	err := c.Bind(&entryFilter)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	entries, total, toast := entry_service.FilterEntries(entryFilter, page, farm)
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
	}

	if len(entries) == 0 {
		c.HTML(http.StatusOK, "no-entry-found-for-filter", gin.H{})
		return
	}

	pageSize := 10
	totalPages := (total + pageSize - 1) / pageSize

	c.HTML(http.StatusOK, "entry-table", gin.H{
		"Entries":     entries,
		"TotalPages":  totalPages,
		"CurrentPage": page,
		"HasPrev":     page > 1,
		"PrevPage":    page - 1,
		"HasNext":     page < totalPages,
		"NextPage":    page + 1,
	})
}

func getEntryFiltersForm(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	c.HTML(http.StatusOK, "entry-filter-form", entry_view.GetFiltersForm(farm))
}

func getEmptyEntryForm(c *gin.Context) {
	draftIdStr := c.Query("draft_id")
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)

	var formMembers entry_view.EntryForm
	var toasts []*entity_public.Toast

	if draftIdStr != "" {
		draftId, err := strconv.ParseUint(draftIdStr, 10, 32)
		if err != nil {
			toast := entity_public.GetWarningToast("ID de rascunho inválido", "")
			c.Header("HX-Trigger", string(toast.ToJson()))
			c.Status(http.StatusBadRequest)
			return
		}
		formMembers, toasts = entry_view.GetEntryFormFromDraft(uint32(draftId), farm)
	} else {
		formMembers, toasts = entry_view.GetEntryForm(farm)
	}

	for _, t := range toasts {
		if t != nil {
			c.Header("HX-Trigger", string(t.ToJson()))
		}
	}
	nonce, _ := c.Get("csp_nonce")
	formMembers.CSPNonce = nonce.(string)
	c.HTML(http.StatusOK, "entry-form", formMembers)
}

func getEntryPdf(c *gin.Context) {
	idPathParam := c.Param("id")
	id, err := strconv.ParseUint(idPathParam, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}

	entryPdf, t := entry_service.GetEntryPdf(uint32(id))
	if err != nil {
		c.Header("HX-Trigger", string(t.ToJson()))
		return
	}
	if entryPdf == nil {
		notFoundToast := entity_public.GetInfoToast("Entrada não encontrada", "")
		c.Header("HX-Trigger", notFoundToast.ToJsonStr())
		c.Status(http.StatusNoContent)
		return
	}
	c.HTML(200, "entry-pdf", entryPdf)
}

func getEntryDraftForm(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	formMembers, _ := entry_view.GetEntryDraftForm(farm)

	nonce, _ := c.Get("csp_nonce")
	formMembers.CSPNonce = nonce.(string)
	c.HTML(http.StatusOK, "entry-draft-form", formMembers)
}

func deleteDraft(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	toast := entry_service.DeleteEntryDraft(uint32(converted))
	c.Header("HX-Trigger", string(toast.ToJson()))
	c.Status(http.StatusOK)
}

func getFilledEntryDraftForm(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}

	draft, toast := entry_service.GetEntryDraft(uint32(converted))
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	formMembers, _ := entry_view.GetEntryDraftForm(farm)
	formMembers.Draft = draft
	formMembers.SelectedVehicle = draft.Vehicle
	formMembers.SelectedCrop = draft.Crop
	formMembers.SelectedField = draft.Field

	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	nonce, _ := c.Get("csp_nonce")
	formMembers.CSPNonce = nonce.(string)
	c.HTML(http.StatusOK, "entry-draft-form", formMembers)
}

func getEntryDraftTable(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	c.HTML(http.StatusOK, "entry-draft-table", entry_view.GetEntryDraftTable(farm))
}

func UseEntryRoutes(router *gin.Engine) {
	router.GET("/romaneio", getRomaneioPage)
	router.GET("/entry/list", getEntryContent)
	router.GET("/entry/filters", getEntryFiltersForm)
	router.GET("/entry/form", getEmptyEntryForm)
	router.GET("/entry/form/:id", getEntryForm)
	router.GET("/entry/draft/form", getEntryDraftForm)
	router.GET("/entry/draft/form/:id", getFilledEntryDraftForm)
	router.GET("/entry/draft/list", getEntryDraftTable)
	router.POST("/entry", addEntry)
	router.POST("/entry/draft", addEntryDraft)
	router.PUT("/entry/:id", putEntry)
	router.PUT("/entry/draft/:id", putEntryDraft)
	router.DELETE("/entry/:id", deleteEntry)
	router.DELETE("/entry/draft/:id", deleteDraft)
	router.POST("/entry/filter", filterEntries)
	router.GET("/entry/pdf/:id", getEntryPdf)
}
