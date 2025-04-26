package vehicle_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/service/user_service"
	"armazenda/service/vehicle_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type VehicleForm struct {
	Name  string `form:"name"`
	Plate string `form:"plate" binding:"required"`
}

func getVehiclesForm(c *gin.Context) {
	//vehicles, _ := vehicle_service.GetVehicles()
	//c.HTML(http.StatusOK, "vehicle-form", vehicles)
	c.HTML(http.StatusOK, "vehicle-form", nil)
}

func addVehicle(c *gin.Context) {
	var newVehicle VehicleForm
	err := c.Bind(&newVehicle)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	ssi, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(ssi)
	vehicle, addErr := vehicle_service.AddVehicle(entity_public.Vehicle{
		Name:  newVehicle.Name,
		Plate: newVehicle.Plate,
		Farm:  farm,
	})

	if addErr != nil {
		t := entity_public.GetWarningToast(addErr.Error(), "")
		c.Header("HX-Trigger", string(t.ToJson()))
		c.Status(http.StatusBadRequest)
		return
	}

	t := entity_public.GetSuccessToast("Veículo Cadastrado", "")
	c.Header("HX-Trigger", string(t.ToJson()))
	c.HTML(http.StatusCreated, "vehicle-option", vehicle)
}

func UseVehicleRouter(router *gin.Engine) {
	router.GET("/vehicle/form", getVehiclesForm)
	router.POST("/vehicle", addVehicle)
}
