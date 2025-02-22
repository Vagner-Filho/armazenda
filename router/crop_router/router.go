package crop_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/crop_model"
	crop_view "armazenda/view/crop"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CropForm struct {
	Name      string `form:"name"`
	Product   uint8  `form:"product"`
	StartDate string `form:"startDate"`
}

func GetCropForm(c *gin.Context) {
	cropForm, toast := crop_view.GetCropForm()
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
	}
	c.HTML(http.StatusOK, "crop-form", cropForm)
}

func AddCrop(c *gin.Context) {
	var newCrop CropForm
	err := c.Bind(&newCrop)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	startDateTime, startDateErr := time.Parse("2006-01-02", newCrop.StartDate)
	if startDateErr != nil {
		c.String(http.StatusBadRequest, "", startDateErr.Error())
		return
	}

	cropModel, _ := crop_model.GetCropModel()
	addedCrop, addErr := cropModel.AddCrop(entity_public.Crop{
		Name:      newCrop.Name,
		StartDate: startDateTime,
		Product:   newCrop.Product,
	})

	if addErr != nil {
		if addErr.IsServerErr == true {
			toast := entity_public.GetErrorToast(addErr.Error(), "")
			c.Header("HX-Trigger", string(toast.ToJson()))
			return
		}

		t := entity_public.GetWarningToast(addErr.Error(), "")
		c.Header("HX-Trigger", string(t.ToJson()))
		c.Status(http.StatusBadRequest)
		return
	}

	t := entity_public.GetSuccessToast("Safra Cadastrada", "")
	c.Header("HX-Trigger", string(t.ToJson()))
	c.HTML(http.StatusCreated, "crop-option", addedCrop)
}
