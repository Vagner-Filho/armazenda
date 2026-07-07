package nfe_router

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	entity_public "armazenda/entity/public"
	"armazenda/model/departure_model"
	"armazenda/model/nfe_model"
	"armazenda/model/product_model"
	"armazenda/pkg/nfe/entity"
	nfe_pdf "armazenda/pkg/nfe/service"
	nfe_xml "armazenda/pkg/nfe/xml"
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

	cfop := c.PostForm("cfop")

	userRates, rateErr := parseUserTaxRates(c)
	if rateErr != nil {
		toast := entity_public.GetWarningToast("Falhamos ao identificar as alíquotas", rateErr.Error())
		c.Header("HX-Trigger", string(toast.ToJson()))
		c.Status(http.StatusBadRequest)
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)

	service := nfe_service.NewNFeService()
	signedXML, toast := service.BuildInvoiceFromDeparture(uint32(departureID), unitPrice, farm, cfop, userRates)

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

func previewNFe(c *gin.Context) {
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

	cfop := c.PostForm("cfop")

	userRates, rateErr := parseUserTaxRates(c)
	if rateErr != nil {
		toast := entity_public.GetWarningToast("Falhamos ao identificar as alíquotas", rateErr.Error())
		c.Header("HX-Trigger", string(toast.ToJson()))
		c.Status(http.StatusBadRequest)
		return
	}

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)

	svc := nfe_service.NewNFeService()
	pdfBytes, toast := svc.GeneratePreviewDANFE(uint32(departureID), unitPrice, farm, cfop, userRates)
	if toast.Type == entity_public.ErrorToast || toast.Type == entity_public.WarningToast {
		c.Header("HX-Trigger", string(toast.ToJson()))
		if toast.Type == entity_public.WarningToast {
			c.Status(http.StatusBadRequest)
		} else {
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	dModel := departure_model.GetDepartureModel()
	departure, depErr := dModel.GetDeparture(uint32(departureID))
	if depErr != nil {
		c.String(http.StatusInternalServerError, "Failed to get departure")
		return
	}

	pModel := product_model.GetProductModel()
	product, prodErr := pModel.GetProductById(departure.Crop)
	productName := "Produto"
	if prodErr == nil {
		productName = product.Name
	}

	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "nfe-preview", gin.H{
		"DepartureID": departureID,
		"Product":     productName,
		"NetWeight":   departure.NetWeight,
		"UnitPrice":   unitPrice,
		"TotalValue":  departure.NetWeight.Mul(unitPrice),
		"PDFBase64":   base64.StdEncoding.EncodeToString(pdfBytes),
		"CFOP":        cfop,
		"ICMSRate":    rateDisplayString(userRates.ICMSRate),
		"PISRate":     rateDisplayString(userRates.PISRate),
		"COFINSRate":  rateDisplayString(userRates.COFINSRate),
		"CSPNonce":    nonce.(string),
	})
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
		"FarmID":        farmID,
		"Config":        config,
		"ICMSRatePct":   percentDisplay(config.ICMSRate),
		"PISRatePct":    percentDisplay(config.PISRate),
		"COFINSRatePct": percentDisplay(config.COFINSRate),
	})
}

func saveNFeConfig(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	// Parse form fields
	environment, _ := strconv.Atoi(c.PostForm("environment"))
	environment = 2
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

	// Tax config defaults
	defaultCFOP := c.PostForm("defaultCFOP")
	if defaultCFOP == "" {
		defaultCFOP = "5101"
	}
	defaultUnit := c.PostForm("defaultUnit")
	if defaultUnit == "" {
		defaultUnit = "KG"
	}
	defaultCEST := strings.TrimSpace(c.PostForm("defaultCEST"))
	defaultICMSCST := strings.TrimSpace(c.PostForm("defaultICMSCST"))
	defaultPISCST := strings.TrimSpace(c.PostForm("defaultPISCST"))
	defaultCOFINSCST := strings.TrimSpace(c.PostForm("defaultCOFINSCST"))
	defaultNaturezaOp := strings.TrimSpace(c.PostForm("defaultNaturezaOp"))

	icmsRate, icmsRateErr := parsePercentRateOrNil(c.PostForm("icmsRate"))
	if icmsRateErr != nil {
		c.HTML(http.StatusBadRequest, "nfe-config-form", gin.H{
			"FarmID": farmID,
			"Error":  "Alíquota ICMS inválida: " + icmsRateErr.Error(),
		})
		return
	}
	pisRate, pisRateErr := parsePercentRateOrNil(c.PostForm("pisRate"))
	if pisRateErr != nil {
		c.HTML(http.StatusBadRequest, "nfe-config-form", gin.H{
			"FarmID": farmID,
			"Error":  "Alíquota PIS inválida: " + pisRateErr.Error(),
		})
		return
	}
	cofinsRate, cofinsRateErr := parsePercentRateOrNil(c.PostForm("cofinsRate"))
	if cofinsRateErr != nil {
		c.HTML(http.StatusBadRequest, "nfe-config-form", gin.H{
			"FarmID": farmID,
			"Error":  "Alíquota COFINS inválida: " + cofinsRateErr.Error(),
		})
		return
	}

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
		CertificateData:              certData,
		CertificatePasswordEncrypted: encryptedPassword,
		Environment:                  environment,
		Serie:                        serie,
		TaxRegime:                    taxRegime,
		EmitterType:                  emitterType,
		IEEmitter:                    ieEmitter,
		EmitterUF:                    "MT", // Default for now
		DefaultModFrete:              defaultModFrete,
		DefaultCFOP:                  defaultCFOP,
		DefaultUnit:                  defaultUnit,
		ICMSRate:                     derefDecimal(icmsRate),
		PISRate:                      derefDecimal(pisRate),
		COFINSRate:                   derefDecimal(cofinsRate),
	}

	if cnpjEmitter != "" {
		config.CNPJEmitter = &cnpjEmitter
	}
	if cpfEmitter != "" {
		config.CPFEmitter = &cpfEmitter
	}
	if defaultCEST != "" {
		config.DefaultCEST = &defaultCEST
	}
	if defaultICMSCST != "" {
		config.DefaultICMSCST = &defaultICMSCST
	}
	if defaultPISCST != "" {
		config.DefaultPISCST = &defaultPISCST
	}
	if defaultCOFINSCST != "" {
		config.DefaultCOFINSCST = &defaultCOFINSCST
	}
	if defaultNaturezaOp != "" {
		config.DefaultNaturezaOp = &defaultNaturezaOp
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

// derefDecimal returns the decimal value pointed to by d, or zero if d is nil.
func derefDecimal(d *decimal.Decimal) decimal.Decimal {
	if d == nil {
		return decimal.Zero
	}
	return *d
}

func UseNFeRoutes(router *gin.Engine) {
	router.GET("/nfe", getNFePage)
	router.POST("/nfe/preview/:departureId", previewNFe)
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

	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}

	nfeModel := nfe_model.GetNFeModel()
	invoices, total, err := nfeModel.GetInvoicesByFarm(farmID, page)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to get invoices")
		return
	}

	pageSize := 10
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	c.HTML(http.StatusOK, "nfe-list", gin.H{
		"Invoices":    invoices,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"HasPrev":     page > 1,
		"PrevPage":    page - 1,
		"HasNext":     page < totalPages,
		"NextPage":    page + 1,
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

	// DANFE can ONLY be generated for authorized invoices (legal requirement)
	if invoice.Status != "authorized" {
		c.String(http.StatusForbidden, "DANFE somente disponivel para NF-e autorizada pela SEFAZ")
		return
	}

	var xmlContent string
	if invoice.XMLAuthorized != nil && *invoice.XMLAuthorized != "" {
		xmlContent = *invoice.XMLAuthorized
	} else if invoice.XMLSigned != nil && *invoice.XMLSigned != "" {
		xmlContent = *invoice.XMLSigned
	} else {
		c.String(http.StatusNotFound, "Invoice XML not available")
		return
	}

	data, parseErr := nfe_xml.ParseDANFEData(xmlContent)
	if parseErr != nil {
		c.String(http.StatusInternalServerError, "Failed to parse invoice data")
		return
	}

	// Fallback: if the XML doesn't contain <protNFe> (e.g., older invoices
	// authorized before the <nfeProc> wrapper was stored), populate the
	// protocol from the database column so Campo 2 on the DANFE is filled.
	if data.Protocol == "" && invoice.Protocol != nil {
		data.Protocol = *invoice.Protocol
	}

	generator := nfe_pdf.NewDANFEGenerator()
	pdfBytes, genErr := generator.Generate(*data)
	if genErr != nil {
		c.String(http.StatusInternalServerError, "Failed to generate DANFE")
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=danfe_%s.pdf", accessKey))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
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

	// Check if an invoice already exists for this departure
	nfeModel := nfe_model.GetNFeModel()
	invoice, invErr := nfeModel.GetInvoiceByDeparture(uint32(departureID))
	if invErr != nil {
		c.String(http.StatusInternalServerError, "Failed to check existing invoice")
		return
	}

	nonce, _ := c.Get("csp_nonce")
	if invoice != nil {
		c.HTML(http.StatusOK, "nfe-existing-modal", gin.H{
			"DepartureID": departureID,
			"Invoice":     invoice,
			"CSPNonce":    nonce.(string),
		})
		return
	}

	// Get product name
	pModel := product_model.GetProductModel()
	product, prodErr := pModel.GetProductByCrop(departure.Crop)
	if prodErr != nil {
		product.Name = "Produto"
	}

	// Load the farm's NFe config to pre-populate the rate inputs with the
	// configured defaults. Missing config is non-fatal — the form will simply
	// show empty defaults.
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)
	var defaultICMS, defaultPIS, defaultCOFINS decimal.Decimal
	if fc, fcErr := nfeModel.GetFarmConfig(farmID); fcErr == nil && fc != nil {
		defaultICMS = fc.ICMSRate
		defaultPIS = fc.PISRate
		defaultCOFINS = fc.COFINSRate
	}

	c.HTML(http.StatusOK, "nfe-emit-modal", gin.H{
		"DepartureID":       departureID,
		"Product":           product.Name,
		"NetWeight":         departure.NetWeight,
		"DefaultICMSRate":   percentDisplay(defaultICMS),
		"DefaultPISRate":    percentDisplay(defaultPIS),
		"DefaultCOFINSRate": percentDisplay(defaultCOFINS),
		"CSPNonce":          nonce.(string),
	})
}

// parsePercentRateOrNil converts a form-submitted percentage string (e.g. "17,00")
// into a decimal rate (0.17) wrapped in a pointer. An empty string yields nil,
// which the service interprets as "user did not provide a value" and falls back
// to the product config. A non-empty string is divided by 100 to produce the
// decimal rate stored in the DB and on the invoice row.
func parsePercentRateOrNil(s string) (*decimal.Decimal, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return nil, err
	}
	if d.IsNegative() {
		return nil, fmt.Errorf("rate must be non-negative")
	}
	rate := d.Div(decimal.NewFromInt(100))
	return &rate, nil
}

// parseUserTaxRates extracts the three rate fields from a form submission.
// An error from any of the three field parses is returned (the caller converts
// it into a 400 response). Empty fields are tolerated and yield nil pointers.
func parseUserTaxRates(c *gin.Context) (entity.TaxRates, error) {
	icms, err := parsePercentRateOrNil(c.PostForm("icmsRate"))
	if err != nil {
		return entity.TaxRates{}, fmt.Errorf("ICMS: %w", err)
	}
	pis, err := parsePercentRateOrNil(c.PostForm("pisRate"))
	if err != nil {
		return entity.TaxRates{}, fmt.Errorf("PIS: %w", err)
	}
	cofins, err := parsePercentRateOrNil(c.PostForm("cofinsRate"))
	if err != nil {
		return entity.TaxRates{}, fmt.Errorf("COFINS: %w", err)
	}
	return entity.TaxRates{ICMSRate: icms, PISRate: pis, COFINSRate: cofins}, nil
}

// rateDisplayString converts an optional decimal rate back to the percentage
// string used in form fields (e.g. 0.17 → "17.00"). Nil pointers yield empty
// strings so the form starts blank for "not provided" cases.
func rateDisplayString(rate *decimal.Decimal) string {
	if rate == nil {
		return ""
	}
	return percentDisplay(*rate)
}

// percentDisplay formats a decimal rate (e.g. 0.17) for display in the form,
// keeping two decimal places of precision.
func percentDisplay(rate decimal.Decimal) string {
	if rate.IsZero() {
		return ""
	}
	pct := rate.Mul(decimal.NewFromInt(100))
	// Use a fixed two-decimal representation; the shopspring/decimal String()
	// method preserves trailing zeros for the value but does not pad.
	return pct.StringFixed(2)
}
