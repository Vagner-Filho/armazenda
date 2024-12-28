package departure_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/buyer_model"
	"armazenda/model/departure_model"
	"armazenda/model/vehicle_model"
	"armazenda/router/vehicle_router"
	"armazenda/service/departure_service"
	"armazenda/service/vehicle_service"
	"armazenda/view/departure"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetDepartureContent(c *gin.Context) {
	c.HTML(http.StatusOK, "departure-content", gin.H{
		"Departures": departure_view.GetDepartures(),
	})
}

func GetDepartureForm(c *gin.Context) {
	c.HTML(http.StatusOK, "departure-form", gin.H{
		"Vehicles": vehicle_service.GetVehicles(),
		"Buyers":   buyer_model.GetBuyers(),
	})
}

type FilledDeparture struct {
	entity_public.Departure
	Vehicles []vehicle_model.Vehicle
	Buyers   []entity_public.Buyer
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

	var vehicles []vehicle_model.Vehicle
	for _, vehicle := range vehicle_router.GetVehicles() {
		vehicles = append(vehicles, vehicle_model.Vehicle{
			Selected: departure.VehiclePlate == vehicle.Plate,
			Plate:    vehicle.Plate,
			Name:     vehicle.Name,
		})
	}

	var buyers []entity_public.Buyer
	for _, buyer := range buyer_model.GetBuyers() {
		buyers = append(buyers, entity_public.Buyer{
			Selected: departure.Buyer == buyer.Id,
			Name:     buyer.Name,
			Id:       buyer.Id,
		})
	}
	filledDeparture := FilledDeparture{
		Departure: departure,
		Vehicles:  vehicles,
		Buyers:    buyers,
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
	c.HTML(http.StatusOK, "departure-list-item", departure_view.MakeReadableDeparture(newDeparture))
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
	if notFound {
		// handle not found
	}

	c.HTML(http.StatusOK, "departure-list-item", departure_view.MakeReadableDeparture(updatedDeparture))
}

func DeleteDeparture(c *gin.Context) {
	id := c.Param("id")
	converted, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
	}

	c.String(http.StatusOK, "", departure_service.DeleteDeparture(uint32(converted)))
}

func FilterDepartures(c *gin.Context) {
	var departureFilter entity_public.DepartureFilter
	err := c.Bind(&departureFilter)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	rawDepartures, err := departure_model.FilterDepartures(departureFilter)

	if err != nil {
		c.HTML(http.StatusBadRequest, "toast", err.Error())
		return
	}

	if len(rawDepartures) == 0 {
		c.HTML(http.StatusOK, "no-departure-found-for-filter", gin.H{})
		return
	}

	var simpleDepartures []departure_view.ReadableDeparture

	for _, departure := range rawDepartures {
		simpleDepartures = append(simpleDepartures, departure_view.MakeReadableDeparture(departure))
	}

	c.HTML(http.StatusOK, "departure-table", simpleDepartures)
}
