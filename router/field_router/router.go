package field_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/field_model"
	"armazenda/service/user_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FieldForm struct {
	Name string `form:"name" binding:"required"`
	Id   uint32 `form:"id"`
}

func getFieldForm(c *gin.Context) {
	fields := []entity_public.Field{}
	var regexPattern string = "^(?!"
	for i, field := range fields {
		regexPattern += field.Name + "$"
		if i < len(fields)-1 {
			regexPattern += "|"
		}
	}
	regexPattern += ").*"
	c.HTML(http.StatusOK, "field-form", nil)
}

func addField(c *gin.Context) {
	var newField FieldForm
	err := c.Bind(&newField)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	ssi, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(ssi)
	fieldModel, _ := field_model.GetFieldModel()
	addedField, addErr := fieldModel.AddField(entity_public.Field{
		Name: newField.Name,
		Farm: farm,
	})

	if addErr != nil {
		if addErr.IsServerErr == true {
			t := entity_public.GetErrorToast(addErr.Error(), "")
			c.Header("HX-Trigger", string(t.ToJson()))
			return
		}

		t := entity_public.GetWarningToast(addErr.Error(), "")
		c.Header("HX-Trigger", string(t.ToJson()))
		c.Status(http.StatusBadRequest)
		return
	}

	t := entity_public.GetSuccessToast("Talhão Cadastrado", "")
	c.Header("HX-Trigger", string(t.ToJson()))
	c.HTML(http.StatusCreated, "field-option", addedField)
}

func UseFieldRoutes(router *gin.Engine) {
	router.POST("/field", addField)
	router.GET("/entry/field/form", getFieldForm)
}
