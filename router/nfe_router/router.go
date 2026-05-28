package nfe_router

import (
	"net/http"
	"strconv"

	entity_public "armazenda/entity/public"
	"armazenda/service/nfe_service"
	"armazenda/service/user_service"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func buildNFe(c *gin.Context) {
	departureIDStr := c.Param("departureId")
	departureID, err := strconv.ParseUint(departureIDStr, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid departure ID")
		return
	}

	unitPriceStr := c.PostForm("unitPrice")
	unitPrice, err := decimal.NewFromString(unitPriceStr)
	if err != nil {
		toast := entity_public.GetWarningToast("Invalid unit price", "")
		c.Header("HX-Trigger", string(toast.ToJson()))
		c.Status(http.StatusBadRequest)
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)

	service := nfe_service.NewNFeService()
	signedXML, toast := service.BuildInvoiceFromDeparture(uint32(departureID), unitPrice, farm)

	if toast.Type == entity_public.ErrorToast || toast.Type == entity_public.WarningToast {
		c.Header("HX-Trigger", string(toast.ToJson()))
		if toast.Type == entity_public.WarningToast {
			c.Status(http.StatusBadRequest)
		} else {
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	c.Header("HX-Trigger", string(toast.ToJson()))
	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, signedXML)
}

func UseNFeRoutes(router *gin.Engine) {
	router.POST("/nfe/build/:departureId", buildNFe)
}
