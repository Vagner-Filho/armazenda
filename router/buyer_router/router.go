package buyer_router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddBuyerCompany(c *gin.Context) {
    var newCompany Company
	err := c.Bind(&newCompany)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}
}

func AddBuyerPerson(c *gin.Context) {
var newPersonal Personal 
	err := c.Bind(&newPersonal)
	if err != nil {
		c.String(http.StatusBadRequest, "", err.Error())
		return
	}
}
