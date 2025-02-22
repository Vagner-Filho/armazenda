package buyer_router

import (
	entity_public "armazenda/entity/public"
	buyer_service "armazenda/service/buyer"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddBuyerCompany(c *gin.Context) {
	var newCompany entity_public.BuyerCompany
	err := c.Bind(&newCompany)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	buyer, toast := buyer_service.AddBuyerCompany(newCompany)
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
	}
	c.HTML(http.StatusCreated, "buyer-option", buyer)
}

func AddBuyerPerson(c *gin.Context) {
	var newPersonal entity_public.BuyerPerson
	err := c.Bind(&newPersonal)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	buyer, toast := buyer_service.AddBuyerPerson(newPersonal)
	if toast != nil {
		c.Header("HX-Trigger", string(toast.ToJson()))
	}
	c.HTML(http.StatusOK, "buyer-option", buyer)
}

func GetBuyerForm(c *gin.Context) {
	c.HTML(http.StatusOK, "buyer-form", gin.H{})
}
