package departure_router

import (
	"armazenda/model/departure_model"
	"armazenda/model/entry_model"
	"armazenda/model/vehicle_model"
	"armazenda/router/entry_router"
	"armazenda/router/vehicle_router"
	"armazenda/service/departure_service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetDepartures(c *gin.Context) {
	c.HTML(http.StatusOK, "departure-table", departure_service.GetDepartures())
}

func GetDepartureForm(c *gin.Context) {
	var vehicles []entry_router.Vehicle
	for _, vehicle := range vehicle_router.GetVehicles() {
		newV := entry_router.Vehicle{}
		newV.Name = vehicle.Name
		newV.Plate = vehicle.Plate
		vehicles = append(vehicles, newV)
	}
	c.HTML(http.StatusOK, "departure-form", gin.H{
		"Vehicles": vehicles,
	})
}

type FilledDeparture struct {
	departure_model.Departure
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

type DepartureForm struct {
	Manifest      uint32            `form:"manifest"`
	DepartureDate int64             `form:"departureDate" binding:"required"`
	Product       entry_model.Grain `form:"product"  binding:"gte=0"`
	VehiclePlate  string            `form:"vehiclePlate" binding:"required"`
	Weight        float64           `form:"weight" binding:"required"`
}

func AddDeparture(c *gin.Context) {
	var df DepartureForm
	err := c.Bind(&df)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	bd := departure_model.BaseDeparture{
		Product:       df.Product,
		Weight:        df.Weight,
		VehiclePlate:  df.VehiclePlate,
		DepartureDate: df.DepartureDate,
	}

	var newDeparture = departure_service.AddDeparture(bd)
	c.HTML(http.StatusOK, "departure-list-item", departure_service.MakeReadableDeparture(newDeparture))
}

func PutDeparture(c *gin.Context) {
	id := c.Param("id")
	converted, parseErr := strconv.ParseUint(id, 10, 32)
	if parseErr != nil {
		c.String(http.StatusBadRequest, "", parseErr.Error())
		return
	}

	var df DepartureForm
	err := c.Bind(&df)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}
	df.Manifest = uint32(converted)

	toUpdate := departure_model.Departure{
		Manifest: df.Manifest,
		BaseDeparture: departure_model.BaseDeparture{
			Weight:        df.Weight,
			Product:       df.Product,
			VehiclePlate:  df.VehiclePlate,
			DepartureDate: df.DepartureDate,
		},
	}

	updatedDeparture, notFound := departure_service.PutDeparture(toUpdate)
    fmt.Printf("%+v\n", updatedDeparture)
    if notFound {
        // handle not found
    }

    c.HTML(http.StatusOK, "departure-list-item", departure_service.MakeReadableDeparture(updatedDeparture))
}
