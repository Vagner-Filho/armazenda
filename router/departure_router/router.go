package departure_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/departure_model"
	"armazenda/service/departure_service"
	"armazenda/service/user_service"
	departure_view "armazenda/view/departure"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getDepartureContent(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	content, toasts := departure_view.GetDepartureContent(farm)
	for _, toast := range toasts {
		if toast != nil {
			c.Header("HX-Trigger", string(toast.ToJson()))
		}
	}
	c.HTML(http.StatusOK, "departure-content", content)
}

func getDepartureForm(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	form, toasts := departure_view.GetNewDepartureForm(farm)

	for _, toast := range toasts {
		if toast != nil {
			c.Header("HX-Trigger", string(toast.ToJson()))
		}
	}

	c.HTML(http.StatusOK, "departure-form", form)
}

func getFilledDepartureForm(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	form, toasts := departure_view.GetExistingDepartureForm(uint32(converted), farm)

	for _, t := range toasts {
		if t != nil {
			c.Header("HX-Trigger", string(t.ToJson()))
		}
	}

	c.HTML(http.StatusOK, "departure-form", form)
}

func addDeparture(c *gin.Context) {
	var df entity_public.Departure
	err := c.Bind(&df)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	df.Farm = farm
	departure, toast := departure_service.AddDeparture(df)
	if toast != nil {
		json := string(toast.ToJson())
		fmt.Printf("\n%v\n", json)
		c.Header("HX-Trigger", json)
	}

	c.HTML(http.StatusCreated, "departure-list-item", departure)
}

func putDeparture(c *gin.Context) {
	id := c.Param("id")
	converted, parseErr := strconv.ParseUint(id, 10, 32)
	if parseErr != nil {
		c.String(http.StatusBadRequest, "", parseErr.Error())
		return
	}

	var df entity_public.Departure
	err := c.Bind(&df)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	df.Id = uint32(converted)

	updatedDeparture, toast := departure_service.PutDeparture(df)
	c.Header("HX-Trigger", string(toast.ToJson()))

	c.HTML(http.StatusOK, "departure-list-item", updatedDeparture)
}

func deleteDeparture(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	toast := departure_service.DeleteDeparture(uint32(converted))
	c.Header("HX-Trigger", string(toast.ToJson()))
	c.Status(http.StatusOK)
}

func filterDepartures(c *gin.Context) {
	var departureFilter entity_public.DepartureFilter
	err := c.Bind(&departureFilter)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	dModel := departure_model.GetDepartureModel()

	departures, err := dModel.FilterDepartures(departureFilter)

	if err != nil {
		toast := entity_public.GetWarningToast("Falha ao filtrar saídas", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	if len(departures) == 0 {
		c.HTML(http.StatusOK, "no-departure-found-for-filter", gin.H{})
		return
	}

	c.HTML(http.StatusOK, "departure-table", departures)
}

func getDeparturePdf(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	departurePdf, t := departure_service.GetDeparturePdf(uint32(id))
	if t != nil {
		c.Header("HX-Trigger", string(t.ToJson()))
		return
	}
	if departurePdf == nil {
		notFoundToast := entity_public.GetInfoToast("Romaneio de saída não encontrado", "")
		c.Header("HX-Trigger", string(notFoundToast.ToJson()))
		c.Status(http.StatusNoContent)
		return
	}
	c.HTML(200, "departure-pdf", departurePdf)
}

func getDepartureDraftForm(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	form, toasts := departure_view.GetDepartureDraftForm(farm)
	for _, toast := range toasts {
		if toast != nil {
			c.Header("HX-Trigger", string(toast.ToJson()))
		}
	}
	c.HTML(http.StatusOK, "departure-draft-form", form)
}

func addDepartureDraft(c *gin.Context) {
	var newDraft entity_public.DepartureDraft
	err := c.Bind(&newDraft)
	if err != nil {
		toast := entity_public.GetWarningToast(err.Error(), "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)
	newDraft.Farm = farm
	draft, toast := departure_service.CreateDepartureDraft(newDraft)
	c.Header("HX-Trigger", string(toast.ToJson()))

	if toast.Type == entity_public.WarningToast {
		c.Status(http.StatusBadRequest)
		return
	}
	if toast.Type == entity_public.ErrorToast {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.HTML(http.StatusCreated, "departure-draft-list-item", draft)
}

func putDepartureDraft(c *gin.Context) {
	id := c.Param("id")
	converted, parseErr := strconv.ParseUint(id, 10, 32)
	if parseErr != nil {
		c.String(http.StatusBadRequest, "", parseErr.Error())
		return
	}

	var draft entity_public.DepartureDraft
	err := c.Bind(&draft)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	draft.Id = uint32(converted)
	updatedDraft, toast := departure_service.UpdateDepartureDraft(draft)
	c.Header("HX-Trigger", string(toast.ToJson()))

	if toast.Type == entity_public.WarningToast {
		c.Status(http.StatusBadRequest)
		return
	}
	if toast.Type == entity_public.ErrorToast {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.HTML(http.StatusOK, "departure-draft-list-item", updatedDraft)
}

func deleteDepartureDraft(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}

	toast := departure_service.DeleteDepartureDraft(uint32(converted))
	c.Header("HX-Trigger", string(toast.ToJson()))
	c.Status(http.StatusOK)
}

func UseDepartureRoutes(router *gin.Engine) {
	router.POST("/departure/filter", filterDepartures)
	router.GET("/departure/list", getDepartureContent)
	router.GET("/departure/form", getDepartureForm)
	router.GET("/departure/form/:id", getFilledDepartureForm)
	router.POST("/departure", addDeparture)
	router.PUT("/departure/:id", putDeparture)
	router.DELETE("/departure/:id", deleteDeparture)
	router.GET("/departure/pdf/:id", getDeparturePdf)

	router.GET("/departure/draft/form", getDepartureDraftForm)
	router.POST("/departure/draft", addDepartureDraft)
	router.PUT("/departure/draft/:id", putDepartureDraft)
	router.DELETE("/departure/draft/:id", deleteDepartureDraft)
}
