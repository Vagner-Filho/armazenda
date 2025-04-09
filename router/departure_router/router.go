package departure_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/departure_model"
	"armazenda/service/departure_service"
	"armazenda/view/departure"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetDepartureContent(c *gin.Context) {
	content, toasts := departure_view.GetDepartureContent()
	for _, toast := range toasts {
		if toast != nil {
			c.Header("HX-Trigger", string(toast.ToJson()))
		}
	}
	if len(content.Departures) == 0 {
		c.HTML(http.StatusOK, "no-departure", gin.H{})
		return
	}
	c.HTML(http.StatusOK, "departure-content", content)
}

func GetDepartureForm(c *gin.Context) {
	form, toasts := departure_view.GetNewDepartureForm()

	for _, toast := range toasts {
		if toast != nil {
			c.Header("HX-Trigger", string(toast.ToJson()))
		}
	}

	c.HTML(http.StatusOK, "departure-form", form)
}

func GetFilledDepartureForm(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}

	form, toasts := departure_view.GetExistingDepartureForm(uint32(converted))

	for _, t := range toasts {
		if t != nil {
			c.Header("HX-Trigger", string(t.ToJson()))
		}
	}

	c.HTML(http.StatusOK, "departure-form", form)
}

func AddDeparture(c *gin.Context) {
	var df entity_public.Departure
	err := c.Bind(&df)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	departure, toast := departure_service.AddDeparture(df)
	if toast != nil {
		json := string(toast.ToJson())
		fmt.Printf("\n%v\n", json)
		c.Header("HX-Trigger", json)
	}

	c.HTML(http.StatusOK, "departure-list-item", departure)
}

func PutDeparture(c *gin.Context) {
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

	updatedDeparture, notFound := departure_service.PutDeparture(df)
	if notFound {
		// handle not found
	}

	c.HTML(http.StatusOK, "departure-list-item", updatedDeparture)
}

func DeleteDeparture(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	toast := departure_service.DeleteDeparture(uint32(converted))
	c.Header("HX-Trigger", string(toast.ToJson()))
	c.Status(http.StatusNoContent)
}

func FilterDepartures(c *gin.Context) {
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

func UseDepartureRoutes(router *gin.Engine) {
	router.POST("/departure/filter", FilterDepartures)
	router.GET("/departure/list", GetDepartureContent)
	router.GET("/departure/form", GetDepartureForm)
	router.GET("/departure/form/:id", GetFilledDepartureForm)
	router.POST("/departure", AddDeparture)
	router.PUT("/departure/:id", PutDeparture)
	router.DELETE("/departure/:id", DeleteDeparture)
}
