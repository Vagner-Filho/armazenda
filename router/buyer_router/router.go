package buyer_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/buyer_model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddBuyerCompany(c *gin.Context) {
	var newCompany entity_public.Company
	err := c.Bind(&newCompany)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}
	var nc = buyer_model.AddBuyerCompany(newCompany)
	c.HTML(http.StatusCreated, "buyer-option", nc.GetBuyer())
}

func AddBuyerPerson(c *gin.Context) {
	var newPersonal entity_public.Personal
	err := c.Bind(&newPersonal)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}

	var np = buyer_model.AddBuyerPersonal(newPersonal)
	c.HTML(http.StatusOK, "buyer-option", np.GetBuyer())
}

func GetBuyerForm(c *gin.Context) {
	c.HTML(http.StatusOK, "buyer-form", gin.H{})
}
