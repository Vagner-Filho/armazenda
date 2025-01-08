package crop_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/crop_model"
	"armazenda/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CropForm struct {
	Name      string `form:"name"`
	StartDate string `form:"startDate"`
}

func GetCropForm(c *gin.Context) {
	c.HTML(http.StatusOK, "crop-form", nil)
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
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	addedCrop := crop_model.AddCrop(entity_public.Crop{
		Name:      newCrop.Name,
		StartDate: startDateTime.Format(utils.TimeLayout),
	})

	c.HTML(http.StatusCreated, "crop-option", addedCrop)
}
