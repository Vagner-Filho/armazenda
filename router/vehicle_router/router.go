package vehicle_router

import (
	"armazenda/model/vehicle_model"
	"armazenda/service/vehicle_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type VehicleForm struct {
	Name  string `form:"name"`
	Plate string `form:"plate" binding:"required"`
}

func GetVehiclesForm(c *gin.Context) {
	c.HTML(http.StatusOK, "vehicleForm", vehicle_service.GetVehicles())
}

func AddVehicle(c *gin.Context) {
	var newVehicle VehicleForm
	err := c.Bind(&newVehicle)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

    vehicle, _ := vehicle_service.AddVehicle(vehicle_model.Vehicle{
        Name: newVehicle.Name,
        Plate: newVehicle.Plate,
    })

    c.HTML(http.StatusCreated, "vehicleOption", vehicle)
}

func GetVehicles() []vehicle_model.Vehicle {
	return vehicle_service.GetVehicles()
}
