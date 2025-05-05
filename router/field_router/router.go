package field_router

import (
	entity_public "armazenda/entity/public"
	field_service "armazenda/service/field"
	"armazenda/service/user_service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type FieldForm struct {
	Name     string          `form:"name" binding:"required"`
	Id       uint32          `form:"id"`
	Hectares decimal.Decimal `form:"hectares" binding:"required"`
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
	addedField, toast := field_service.AddField(entity_public.Field{
		Name:     newField.Name,
		Farm:     farm,
		Hectares: newField.Hectares,
	})

	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
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
