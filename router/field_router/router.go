package field_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/field_model"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FieldForm struct {
	Name string `form:"name" binding:"required"`
	Id   uint32 `form:"id"`
}

func GetFieldForm(c *gin.Context) {
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

func AddField(c *gin.Context) {
	var newField FieldForm
	err := c.Bind(&newField)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	fieldModel, _ := field_model.GetFieldModel()
	addedField, addErr := fieldModel.AddField(entity_public.Field{
		Name: newField.Name,
	})

	if addErr != nil {
		if addErr.IsServerErr == true {
			c.HTML(http.StatusInternalServerError, "toast", gin.H{
				"Message": addErr.Error(),
				"IsError": true,
			})
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
