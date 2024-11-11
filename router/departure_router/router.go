package departure_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/vehicle_model"
	"armazenda/router/entry_router"
	"armazenda/router/vehicle_router"
	"armazenda/service/departure_service"
	"armazenda/view/departure"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetDepartures(c *gin.Context) {
	c.HTML(http.StatusOK, "departure-table", departure_service.GetDepartures())
}

func GetDepartureForm(c *gin.Context) {
	c.HTML(http.StatusOK, "departure-form", departure_view.GetNewDepartureForm())
}

type FilledDeparture struct {
	entity_public.Departure
	Vehicles []entry_router.Vehicle
}

func GetFilledDepartureForm(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}

	departure, notFound := departure_service.GetDeparture(uint32(converted))
	if notFound {
		c.HTML(http.StatusBadRequest, "toast", gin.H{})
	}

	var vehicles []entry_router.Vehicle
	for _, vehicle := range vehicle_router.GetVehicles() {
		vehicles = append(vehicles, entry_router.Vehicle{
			Selected: departure.VehiclePlate == vehicle.Plate,
			Vehicle: vehicle_model.Vehicle{
				Plate: vehicle.Plate,
				Name:  vehicle.Name,
			},
		})
	}
	filledDeparture := FilledDeparture{
		Departure: departure,
		Vehicles:  vehicles,
	}

	c.HTML(http.StatusOK, "departure-form", filledDeparture)
}

func AddDeparture(c *gin.Context) {
	var df entity_public.Departure
	err := c.Bind(&df)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	var newDeparture = departure_service.AddDeparture(df)
	c.HTML(http.StatusOK, "departure-list-item", departure_service.MakeReadableDeparture(newDeparture))
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
	df.Manifest = uint32(converted)

	updatedDeparture, notFound := departure_service.PutDeparture(df)
	fmt.Printf("%+v\n", updatedDeparture)
	if notFound {
		// handle not found
	}

	c.HTML(http.StatusOK, "departure-list-item", departure_service.MakeReadableDeparture(updatedDeparture))
}

func DeleteDeparture(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}

	c.String(http.StatusOK, "", departure_service.DeleteDeparture(uint32(converted)))
}
