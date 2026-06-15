package nfe_router

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	entity_public "armazenda/entity/public"
	"armazenda/model/departure_model"
	"armazenda/model/nfe_model"
	"armazenda/model/product_model"
	"armazenda/service/nfe_service"
	"armazenda/service/user_service"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func buildNFe(c *gin.Context) {
	departureIDStr := c.Param("departureId")
	departureID, err := strconv.ParseUint(departureIDStr, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "Falhamos ao identificar o número de Romaneio")
		return
	}

	unitPriceStr := c.PostForm("unitPrice")
	unitPrice, err := decimal.NewFromString(unitPriceStr)
	if err != nil {
		toast := entity_public.GetWarningToast("Falhamos ao identificar o preço unitário", "")
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

func getNFeConfigForm(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	nfeModel := nfe_model.GetNFeModel()
	config, err := nfeModel.GetFarmConfig(farmID)
	if err != nil {
		config = &nfe_model.FarmConfig{}
	}
	if config == nil {
		config = &nfe_model.FarmConfig{}
	}

	c.HTML(http.StatusOK, "nfe-config-form", gin.H{
		"FarmID": farmID,
		"Config": config,
	})
}

func saveNFeConfig(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	// Parse form fields
	environment, _ := strconv.Atoi(c.PostForm("environment"))
	emitterType, _ := strconv.Atoi(c.PostForm("emitterType"))
	taxRegime, _ := strconv.Atoi(c.PostForm("taxRegime"))
	defaultModFrete, _ := strconv.Atoi(c.PostForm("defaultModFrete"))
	serie, _ := strconv.Atoi(c.PostForm("serie"))
	if serie == 0 {
		serie = 1
	}

	ieEmitter := c.PostForm("ieEmitter")
	cnpjEmitter := c.PostForm("cnpjEmitter")
	cpfEmitter := c.PostForm("cpfEmitter")

	// Handle certificate upload
	var certPath string
	var certData []byte
	file, header, err := c.Request.FormFile("certificate")
	if err == nil && file != nil {
		defer file.Close()

		// Read certificate bytes directly into memory
		data, readErr := io.ReadAll(file)
		if readErr != nil {
			c.HTML(http.StatusInternalServerError, "nfe-config-form", gin.H{
				"FarmID": farmID,
				"Error":  "Failed to read certificate",
			})
			return
		}
		certData = data
		certPath = header.Filename
	}

	// Get existing config to preserve certificate data if no new upload
	nfeModel := nfe_model.GetNFeModel()
	existingConfig, _ := nfeModel.GetFarmConfig(farmID)
	if certData == nil && existingConfig != nil {
		certData = existingConfig.CertificateData
		certPath = existingConfig.CertificatePath
	}

	// Handle password encryption
	certPassword := c.PostForm("certificatePassword")
	var encryptedPassword string
	if certPassword != "" {
		encryptedPassword = nfe_service.EncryptPassword(certPassword)
	} else if existingConfig != nil {
		encryptedPassword = existingConfig.CertificatePasswordEncrypted
	}

	config := nfe_model.FarmConfig{
		FarmID:                       int(farmID),
		CertificatePath:              certPath,
		CertificatePasswordEncrypted: encryptedPassword,
		Environment:                  environment,
		Serie:                        serie,
		TaxRegime:                    taxRegime,
		EmitterType:                  emitterType,
		IEEmitter:                    ieEmitter,
		EmitterUF:                    "MT", // Default for now
		DefaultModFrete:              defaultModFrete,
	}

	if cnpjEmitter != "" {
		config.CNPJEmitter = &cnpjEmitter
	}
	if cpfEmitter != "" {
		config.CPFEmitter = &cpfEmitter
	}

	if upsertErr := nfeModel.UpsertFarmConfig(config); upsertErr != nil {
		c.HTML(http.StatusInternalServerError, "nfe-config-form", gin.H{
			"FarmID": farmID,
			"Config": config,
			"Error":  "Failed to save NF-e configuration: " + upsertErr.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "nfe-config-form", gin.H{
		"FarmID":  farmID,
		"Config":  config,
		"Success": "Configuração NF-e salva com sucesso!",
	})
}

func UseNFeRoutes(router *gin.Engine) {
	router.GET("/nfe", getNFePage)
	router.POST("/nfe/build/:departureId", buildNFe)
	router.GET("/nfe/config/form", getNFeConfigForm)
	router.PUT("/nfe/config", saveNFeConfig)
	router.GET("/nfe/modal/:departureId", getNFeEmitModal)
	router.GET("/nfe/list", getNFeList)
	router.GET("/nfe/download/xml/:accessKey", downloadNFeXML)
	router.GET("/nfe/download/danfe/:accessKey", downloadNFeDANFE)
}

func getNFePage(c *gin.Context) {
	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "nfe.html", gin.H{
		"CSPNonce": nonce.(string),
	})
}

func getNFeList(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	nfeModel := nfe_model.GetNFeModel()
	invoices, err := nfeModel.GetInvoicesByFarm(farmID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to get invoices")
		return
	}

	c.HTML(http.StatusOK, "nfe-list", gin.H{
		"Invoices": invoices,
	})
}

func downloadNFeXML(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	accessKey := c.Param("accessKey")
	nfeModel := nfe_model.GetNFeModel()
	invoice, err := nfeModel.GetInvoiceByAccessKey(accessKey)
	if err != nil || invoice == nil {
		c.String(http.StatusNotFound, "Invoice not found")
		return
	}

	// Verify the invoice belongs to the user's farm
	dModel := departure_model.GetDepartureModel()
	departure, depErr := dModel.GetDeparture(invoice.DepartureID)
	if depErr != nil || departure.Farm != farmID {
		c.String(http.StatusForbidden, "Access denied")
		return
	}

	var xmlContent string
	if invoice.XMLAuthorized != nil && *invoice.XMLAuthorized != "" {
		xmlContent = *invoice.XMLAuthorized
	} else if invoice.XMLSigned != nil && *invoice.XMLSigned != "" {
		xmlContent = *invoice.XMLSigned
	} else {
		c.String(http.StatusNotFound, "XML not available")
		return
	}

	c.Header("Content-Type", "application/xml")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=nfe_%s.xml", accessKey))
	c.String(http.StatusOK, xmlContent)
}

func downloadNFeDANFE(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	accessKey := c.Param("accessKey")
	nfeModel := nfe_model.GetNFeModel()
	invoice, err := nfeModel.GetInvoiceByAccessKey(accessKey)
	if err != nil || invoice == nil {
		c.String(http.StatusNotFound, "Invoice not found")
		return
	}

	// Verify the invoice belongs to the user's farm
	dModel := departure_model.GetDepartureModel()
	departure, depErr := dModel.GetDeparture(invoice.DepartureID)
	if depErr != nil || departure.Farm != farmID {
		c.String(http.StatusForbidden, "Access denied")
		return
	}

	// TODO: Generate DANFE PDF using pkg/nfe/service/danfe.go
	// For now, return a placeholder message
	c.String(http.StatusNotImplemented, "DANFE generation not yet implemented")
}

func getNFeEmitModal(c *gin.Context) {
	departureIDStr := c.Param("departureId")
	departureID, err := strconv.ParseUint(departureIDStr, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid departure ID")
		return
	}

	dModel := departure_model.GetDepartureModel()
	departure, depErr := dModel.GetDeparture(uint32(departureID))
	if depErr != nil {
		c.String(http.StatusInternalServerError, "Failed to get departure")
		return
	}

	// Get product name
	pModel := product_model.GetProductModel()
	product, prodErr := pModel.GetProductById(departure.Crop)
	if prodErr != nil {
		product.Name = "Produto"
	}

	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "nfe-emit-modal", gin.H{
		"DepartureID": departureID,
		"Product":     product.Name,
		"NetWeight":   departure.NetWeight,
		"CSPNonce":    nonce.(string),
	})
}
