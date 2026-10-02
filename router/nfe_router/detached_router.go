package nfe_router

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	entity_public "armazenda/entity/public"
	"armazenda/model/nfe_model"
	"armazenda/model/person_model"
	"armazenda/service/nfe_service"
	"armazenda/service/user_service"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// getDetachedNFePage renders the full page for detached NF-e emission
func getDetachedNFePage(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	// Get farm NFe config to check if it's configured
	nfeModel := nfe_model.GetNFeModel()
	farmNFeConfig, _ := nfeModel.GetFarmConfig(farmID)

	// Get list of people for recipient selection
	pModel := person_model.GetPersonModel()
	people, _ := pModel.GetPeopleByFarm(farmID)

	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "nfe-emit.html", gin.H{
		"FarmID":           farmID,
		"People":           people,
		"HasNFEFarmConfig": farmNFeConfig != nil,
		"DefaultCFOP":      "5101",
		"DefaultUnit":      "KG",
		"CSPNonce":         nonce.(string),
		"TierKey":          user_service.GetTierKeyFromContext(c),
	})
}

// buildDetachedNFe handles the form submission to emit a detached NF-e
func buildDetachedNFe(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	// Parse recipient
	recipient, toast := parseDetachedRecipient(c)
	if toast.Type == entity_public.ErrorToast || toast.Type == entity_public.WarningToast {
		c.Header("HX-Trigger", toast.ToJsonStr())
		c.Status(http.StatusBadRequest)
		return
	}

	// Parse items
	items, itemsToast := parseDetachedItems(c)
	if itemsToast.Type == entity_public.ErrorToast || itemsToast.Type == entity_public.WarningToast {
		c.Header("HX-Trigger", itemsToast.ToJsonStr())
		c.Status(http.StatusBadRequest)
		return
	}

	// Parse invoice-level fields
	cfop := c.PostForm("cfop")
	if cfop == "" {
		cfop = "5101"
	}

	// Parse tax rates
	userRates, rateErr := parseUserTaxRates(c)
	if rateErr != nil {
		toast := entity_public.GetWarningToast("Falha ao identificar as alíquotas", rateErr.Error())
		c.Header("HX-Trigger", toast.ToJsonStr())
		c.Status(http.StatusBadRequest)
		return
	}

	// Parse overrides
	overrides := parseInvoiceOverrides(c)

	// Parse vehicle (optional)
	var vehicleID *uint16
	if vehicleStr := c.PostForm("vehicleId"); vehicleStr != "" {
		if vid, err := strconv.ParseUint(vehicleStr, 10, 16); err == nil {
			v := uint16(vid)
			vehicleID = &v
		}
	}

	input := nfe_service.DetachedInvoiceInput{
		FarmID:    farmID,
		Recipient: recipient,
		VehicleID: vehicleID,
		CFOP:      cfop,
		TaxRates:  userRates,
		Overrides: overrides,
		Items:     items,
	}

	svc := nfe_service.NewNFeService()
	signedXML, toast := svc.BuildDetachedInvoice(input)

	if toast.Type == entity_public.ErrorToast || toast.Type == entity_public.WarningToast {
		c.Header("HX-Trigger", toast.ToJsonStr())
		if toast.Type == entity_public.WarningToast {
			c.Status(http.StatusBadRequest)
		} else {
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	c.Header("HX-Trigger", toast.ToJsonStr())
	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, signedXML)
}

// parseDetachedRecipient parses the recipient from the form
func parseDetachedRecipient(c *gin.Context) (nfe_service.DetachedRecipient, entity_public.Toast) {
	// Check if using existing person
	if personIDStr := c.PostForm("recipientPersonId"); personIDStr != "" {
		if pid, err := strconv.ParseUint(personIDStr, 10, 32); err == nil {
			personID := uint32(pid)
			return nfe_service.DetachedRecipient{
				PersonID: &personID,
			}, entity_public.Toast{}
		}
	}

	// Otherwise, parse inline recipient
	name := strings.TrimSpace(c.PostForm("recipientName"))
	document := strings.TrimSpace(c.PostForm("recipientDocument"))
	if name == "" || document == "" {
		return nfe_service.DetachedRecipient{}, entity_public.GetWarningToast(
			"Destinatário incompleto",
			"Informe o nome e o documento do destinatário")
	}

	var ie *string
	if v := strings.TrimSpace(c.PostForm("recipientIE")); v != "" {
		ie = &v
	}

	var street, number, neighborhood, city, state, cep, phone, email *string
	if v := strings.TrimSpace(c.PostForm("recipientStreet")); v != "" {
		street = &v
	}
	if v := strings.TrimSpace(c.PostForm("recipientNumber")); v != "" {
		number = &v
	}
	if v := strings.TrimSpace(c.PostForm("recipientNeighborhood")); v != "" {
		neighborhood = &v
	}
	if v := strings.TrimSpace(c.PostForm("recipientCity")); v != "" {
		city = &v
	}
	if v := strings.TrimSpace(c.PostForm("recipientState")); v != "" {
		state = &v
	}
	if v := strings.TrimSpace(c.PostForm("recipientCEP")); v != "" {
		cep = &v
	}
	if v := strings.TrimSpace(c.PostForm("recipientPhone")); v != "" {
		phone = &v
	}
	if v := strings.TrimSpace(c.PostForm("recipientEmail")); v != "" {
		email = &v
	}

	return nfe_service.DetachedRecipient{
		Name:         name,
		Document:     document,
		IE:           ie,
		Street:       street,
		Number:       number,
		Neighborhood: neighborhood,
		City:         city,
		State:        state,
		CEP:          cep,
		PhoneNumber:  phone,
		Email:        email,
	}, entity_public.Toast{}
}

// parseDetachedItems parses the items from the form
func parseDetachedItems(c *gin.Context) ([]nfe_service.DetachedItemInput, entity_public.Toast) {
	// Parse item count
	itemCountStr := c.PostForm("itemCount")
	itemCount, err := strconv.Atoi(itemCountStr)
	if err != nil || itemCount < 1 {
		return nil, entity_public.GetWarningToast("Nenhum item informado", "Adicione pelo menos um item à NF-e")
	}

	var items []nfe_service.DetachedItemInput
	for i := 0; i < itemCount; i++ {
		prefix := fmt.Sprintf("items[%d].", i)

		// Check if using farm product
		var farmProductID *uint16
		if fpIDStr := c.PostForm(prefix + "farmProductId"); fpIDStr != "" {
			if fpID, err := strconv.ParseUint(fpIDStr, 10, 16); err == nil {
				v := uint16(fpID)
				farmProductID = &v
			}
		}

		// Parse quantity
		quantityStr := c.PostForm(prefix + "quantity")
		quantity, err := decimal.NewFromString(quantityStr)
		if err != nil || quantity.IsZero() {
			return nil, entity_public.GetWarningToast(
				"Quantidade inválida",
				fmt.Sprintf("Item %d: informe uma quantidade válida", i+1))
		}

		// Parse gross weight
		grossWeightStr := c.PostForm(prefix + "grossWeight")
		grossWeight, err := decimal.NewFromString(grossWeightStr)
		if err != nil {
			grossWeight = quantity // default to quantity if not provided
		}

		// Parse unit price
		unitPriceStr := c.PostForm(prefix + "unitPrice")
		unitPrice, err := decimal.NewFromString(unitPriceStr)
		if err != nil || unitPrice.IsZero() {
			return nil, entity_public.GetWarningToast(
				"Preço unitário inválido",
				fmt.Sprintf("Item %d: informe um preço unitário válido", i+1))
		}

		// Parse optional fields
		productName := strings.TrimSpace(c.PostForm(prefix + "productName"))
		ncm := strings.TrimSpace(c.PostForm(prefix + "ncm"))
		unit := strings.TrimSpace(c.PostForm(prefix + "unit"))
		if unit == "" {
			unit = "KG"
		}

		var cest *string
		if v := strings.TrimSpace(c.PostForm(prefix + "cest")); v != "" {
			cest = &v
		}

		items = append(items, nfe_service.DetachedItemInput{
			FarmProductID: farmProductID,
			ProductName:   productName,
			NCM:           ncm,
			CEST:          cest,
			Quantity:      quantity,
			GrossWeight:   grossWeight,
			UnitPrice:     unitPrice,
			Unit:          unit,
		})
	}

	return items, entity_public.Toast{}
}

// getDetachedNFeList returns the list of detached invoices
func getDetachedNFeList(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}

	nfeModel := nfe_model.GetNFeModel()
	invoices, total, err := nfeModel.GetDetachedInvoicesByFarm(farmID, page)
	if err != nil {
		c.String(http.StatusInternalServerError, "Falha ao buscar NF-e avulsas")
		return
	}

	pageSize := 10
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	c.HTML(http.StatusOK, "nfe-detached-list", gin.H{
		"Invoices":    invoices,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"HasPrev":     page > 1,
		"PrevPage":    page - 1,
		"HasNext":     page < totalPages,
		"NextPage":    page + 1,
	})
}

// downloadDetachedNFeXML downloads the XML of a detached invoice
func downloadDetachedNFeXML(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	accessKey := c.Param("accessKey")
	nfeModel := nfe_model.GetNFeModel()
	invoice, err := nfeModel.GetDetachedInvoiceByAccessKey(accessKey)
	if err != nil || invoice == nil {
		c.String(http.StatusNotFound, "NF-e não encontrada")
		return
	}

	if invoice.FarmID != farmID {
		c.String(http.StatusForbidden, "Acesso negado")
		return
	}

	var xmlContent string
	if invoice.XMLAuthorized != nil && *invoice.XMLAuthorized != "" {
		xmlContent = *invoice.XMLAuthorized
	} else if invoice.XMLSigned != nil && *invoice.XMLSigned != "" {
		xmlContent = *invoice.XMLSigned
	} else {
		c.String(http.StatusNotFound, "XML não disponível")
		return
	}

	c.Header("Content-Type", "application/xml")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=nfe_%s.xml", accessKey))
	c.String(http.StatusOK, xmlContent)
}

// downloadDetachedNFeDANFE downloads the DANFE of a detached invoice
func downloadDetachedNFeDANFE(c *gin.Context) {
	// TODO: Implement DANFE generation for detached invoices
	// This requires parsing the XML and generating a PDF
	c.String(http.StatusNotImplemented, "DANFE para NF-e avulsa ainda não implementado")
}

// cancelDetachedNFe cancels a detached invoice
func cancelDetachedNFe(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	accessKey := c.Param("accessKey")
	justification := c.PostForm("justification")

	svc := nfe_service.NewNFeService()
	toast := svc.CancelDetachedInvoice(accessKey, justification, farmID)

	if toast.Type == entity_public.ErrorToast || toast.Type == entity_public.WarningToast {
		c.Header("HX-Trigger", toast.ToJsonStr())
		if toast.Type == entity_public.WarningToast {
			c.Status(http.StatusBadRequest)
		} else {
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	c.Header("HX-Trigger", toast.ToJsonStr())
	c.Status(http.StatusOK)
}

// UseDetachedNFeRoutes registers the detached NF-e routes
func UseDetachedNFeRoutes(router gin.IRoutes) {
	router.GET("/emitir", getDetachedNFePage)
	router.POST("/emitir/build", buildDetachedNFe)
	router.GET("/avulsa/list", getDetachedNFeList)
	router.GET("/avulsa/download/xml/:accessKey", downloadDetachedNFeXML)
	router.GET("/avulsa/download/danfe/:accessKey", downloadDetachedNFeDANFE)
	router.POST("/avulsa/cancel/:accessKey", cancelDetachedNFe)
}
