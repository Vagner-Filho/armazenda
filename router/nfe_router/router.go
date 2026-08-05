package nfe_router

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	entity_public "armazenda/entity/public"
	"armazenda/model/departure_model"
	"armazenda/model/farm_config_model"
	"armazenda/model/nfe_model"
	"armazenda/model/person_model"
	"armazenda/model/product_model"
	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"
	nfe_pdf "armazenda/pkg/nfe/service"
	nfe_xml "armazenda/pkg/nfe/xml"
	"armazenda/service/nfe_service"
	"armazenda/service/user_service"
	"armazenda/utils"

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

	overrides := parseInvoiceOverrides(c)

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)

	service := nfe_service.NewNFeService()
	signedXML, toast := service.BuildInvoiceFromDeparture(uint32(departureID), unitPrice, farm, cfop, userRates, overrides)

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

	overrides := parseInvoiceOverrides(c)

	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)

	svc := nfe_service.NewNFeService()
	pdfBytes, toast := svc.GeneratePreviewDANFE(uint32(departureID), unitPrice, farm, cfop, userRates, overrides)
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
		"DepartureID":       departureID,
		"Product":             productName,
		"NetWeight":           departure.NetWeight,
		"UnitPrice":           unitPrice,
		"TotalValue":          departure.NetWeight.Mul(unitPrice),
		"PDFBase64":           base64.StdEncoding.EncodeToString(pdfBytes),
		"CFOP":                cfop,
		"ICMSRate":            rateDisplayString(userRates.ICMSRate),
		"PISRate":             rateDisplayString(userRates.PISRate),
		"COFINSRate":          rateDisplayString(userRates.COFINSRate),
		"NaturezaOp":          safeFormValue(c, "naturezaOp"),
		"ProductDesc":         safeFormValue(c, "productDesc"),
		"NCM":                 safeFormValue(c, "ncm"),
		"CEST":                safeFormValue(c, "cest"),
		"Unit":                safeFormValue(c, "unit"),
		"ModFrete":            safeFormValue(c, "modFrete"),
		"ICMSCST":             safeFormValue(c, "icmsCST"),
		"PISCST":              safeFormValue(c, "pisCST"),
		"COFINSCST":           safeFormValue(c, "cofinsCST"),
		"InfCpl":              safeFormValue(c, "infCpl"),
		"CSPNonce":            nonce.(string),
	})
}

func getNFeConfigForm(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	nfeModel := nfe_model.GetNFeModel()
	config, err := nfeModel.GetFarmConfig(farmID)
	if err != nil {
		config = &entity_public.FarmConfig{}
	}
	if config == nil {
		fcm := farm_config_model.GetFarmConfigModel()
		farm, farmErr := fcm.GetFarmConfig(farmID)
		if farmErr != nil {

		}
		config = &entity_public.FarmConfig{
			IEEmitter:   farm.InscricaoEstadual,
			DocEmitter:  farm.OwnerDocument,
			EmitterType: *farm.OwnerDocumentType,
		}
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

	var fConfig entity_public.FarmConfig
	c.Bind(&fConfig)

	farmID := user_service.GetFarmFromToken(sid)

	fConfig.Environment = 2
	if fConfig.Serie == 0 {
		fConfig.Serie = 1
	}
	if fConfig.DefaultCFOP == "" {
		fConfig.DefaultCFOP = "5101"
	}
	if fConfig.DefaultUnit == "" {
		fConfig.DefaultUnit = "KG"
	}
	if fConfig.DefaultCEST != nil {
		trimmed := strings.TrimSpace(*fConfig.DefaultCEST)
		fConfig.DefaultCEST = &trimmed
	}
	if fConfig.DefaultICMSCST != nil {
		trimmed := strings.TrimSpace(*fConfig.DefaultICMSCST)
		fConfig.DefaultICMSCST = &trimmed
	}
	if fConfig.DefaultPISCST != nil {
		trimmed := strings.TrimSpace(*fConfig.DefaultPISCST)
		fConfig.DefaultPISCST = &trimmed
	}
	if fConfig.DefaultCOFINSCST != nil {
		trimmed := strings.TrimSpace(*fConfig.DefaultCOFINSCST)
		fConfig.DefaultCOFINSCST = &trimmed
	}
	if fConfig.DefaultNaturezaOp != nil {
		trimmed := strings.TrimSpace(*fConfig.DefaultNaturezaOp)
		fConfig.DefaultNaturezaOp = &trimmed
	}

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
	if certData == nil && existingConfig != nil && existingConfig.CertificatePath != nil {
		certData = existingConfig.CertificateData
		certPath = *existingConfig.CertificatePath
	}

	// Handle password encryption
	certPassword := c.PostForm("certificatePassword")
	var encryptedPassword string
	if certPassword != "" {
		encryptedPassword = nfe_service.EncryptPassword(certPassword)
	} else if existingConfig != nil {
		encryptedPassword = existingConfig.CertificatePasswordEncrypted
	}

	config := entity_public.FarmConfig{
		FarmID:                       &farmID,
		CertificatePath:              &certPath,
		CertificateData:              certData,
		CertificatePasswordEncrypted: encryptedPassword,
		Environment:                  fConfig.Environment,
		Serie:                        fConfig.Serie,
		TaxRegime:                    fConfig.TaxRegime,
		DefaultModFrete:              fConfig.DefaultModFrete,
		DefaultCFOP:                  fConfig.DefaultCFOP,
		DefaultUnit:                  fConfig.DefaultUnit,
		ICMSRate:                     derefDecimal(icmsRate),
		PISRate:                      derefDecimal(pisRate),
		COFINSRate:                   derefDecimal(cofinsRate),
		FarmCND: entity_public.FarmCND{
			CertificateNumber: fConfig.CertificateNumber,
			ExpDate:           fConfig.ExpDate,
		},
	}

	meta, empty := utils.ParseMetaFromForm(c)
	if empty == false {
		config.FarmCND.Meta = &meta
	}

	_, toast := nfe_service.UpsertFarmConfig(config)
	if toast.Type == entity_public.ErrorToast {
		c.Header("HX-Trigger", toast.ToJsonStr())
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Header("HX-Trigger", toast.ToJsonStr())
	c.Status(http.StatusOK)
}

// derefDecimal returns the decimal value pointed to by d, or zero if d is nil.
func derefDecimal(d *decimal.Decimal) decimal.Decimal {
	if d == nil {
		return decimal.Zero
	}
	return *d
}

func UseNFeRoutes(router gin.IRoutes) {
	router.GET("/", getNFePage)
	router.POST("/preview/:departureId", previewNFe)
	router.POST("/build/:departureId", buildNFe)
	router.GET("/config/form", getNFeConfigForm)
	router.PUT("/config", saveNFeConfig)
	router.GET("/modal/:departureId", getNFeEmitModal)
	router.GET("/list", getNFeList)
	router.GET("/download/xml/:accessKey", downloadNFeXML)
	router.GET("/download/danfe/:accessKey", downloadNFeDANFE)
	router.GET("/cancel/modal/:accessKey", getNFeCancelModal)
	router.POST("/cancel/:accessKey", cancelNFe)
}

func getNFePage(c *gin.Context) {
	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "nfe.html", gin.H{
		"CSPNonce": nonce.(string),
		"TierKey":  user_service.GetTierKeyFromContext(c),
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

	// DANFE is available for authorized invoices, and for cancelled invoices
	// rendered with an "NF-e CANCELADA" banner per MOC Anexo II.
	if invoice.Status != "authorized" && invoice.Status != "cancelled" {
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
	var pdfBytes []byte
	var genErr error
	if invoice.Status == "cancelled" {
		pdfBytes, genErr = generator.GenerateCancelled(*data)
	} else {
		pdfBytes, genErr = generator.Generate(*data)
	}
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
	var defaultNaturezaOp, defaultCEST, defaultUnit, defaultICMSCST, defaultPISCST, defaultCOFINSCST *string
	var defaultModFrete int
	var defaultCFOP string
	var farmNFeConfig *entity_public.FarmConfig
	hasConfig := false
	if fc, fcErr := nfeModel.GetFarmConfig(farmID); fcErr == nil && fc != nil {
		farmNFeConfig = fc
		defaultICMS = fc.ICMSRate
		defaultPIS = fc.PISRate
		defaultCOFINS = fc.COFINSRate
		defaultNaturezaOp = fc.DefaultNaturezaOp
		defaultCEST = fc.DefaultCEST
		defaultUnit = &fc.DefaultUnit
		if fc.DefaultUnit == "" {
			empty := "KG"
			defaultUnit = &empty
		}
		defaultModFrete = fc.DefaultModFrete
		defaultICMSCST = fc.DefaultICMSCST
		defaultPISCST = fc.DefaultPISCST
		defaultCOFINSCST = fc.DefaultCOFINSCST
		defaultCFOP = fc.DefaultCFOP
		hasConfig = true
	}

	// Derive default natureza op from CFOP if not configured
	if (defaultNaturezaOp == nil || *defaultNaturezaOp == "") && hasConfig {
		derived := defaults.NaturezaOpForCFOP(defaultCFOP)
		defaultNaturezaOp = &derived
	}

	// Build the default infCpl from farm and recipient CND data so the user
	// can review and edit it in the modal.
	var defaultInfCpl string
	if hasConfig && departure.Recipient != nil {
		pModel := person_model.GetPersonModel()
		person, personErr := pModel.GetFullPersonById(*departure.Recipient)
		if personErr == nil {
			defaultInfCpl = nfe_service.BuildDefaultInfCpl(farmNFeConfig.FarmCND, person.PersonCND)
		}
	}

	c.HTML(http.StatusOK, "nfe-emit-modal", gin.H{
		"DepartureID":       departureID,
		"Product":           product.Name,
		"NetWeight":         departure.NetWeight,
		"DefaultICMSRate":   percentDisplay(defaultICMS),
		"DefaultPISRate":    percentDisplay(defaultPIS),
		"DefaultCOFINSRate": percentDisplay(defaultCOFINS),
		"DefaultNaturezaOp": safePtrString(defaultNaturezaOp),
		"DefaultProductDesc": product.Name,
		"DefaultNCM":        product.NCM,
		"DefaultCEST":       safePtrString(defaultCEST),
		"DefaultUnit":       safePtrString(defaultUnit),
		"DefaultModFrete":   defaultModFrete,
		"DefaultICMSCST":    safePtrString(defaultICMSCST),
		"DefaultPISCST":     safePtrString(defaultPISCST),
		"DefaultCOFINSCST":  safePtrString(defaultCOFINSCST),
		"DefaultInfCpl":     defaultInfCpl,
		"CSPNonce":          nonce.(string),
		"HasNFEFarmConfig":  hasConfig,
	})
}

// getNFeCancelModal renders the cancellation confirmation modal for an
// authorized NF-e. It warns when the invoice was authorized more than 24h ago,
// since SEFAZ may reject cancellations past the legal deadline (rejection 501).
func getNFeCancelModal(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	accessKey := c.Param("accessKey")
	nfeModel := nfe_model.GetNFeModel()
	invoice, err := nfeModel.GetInvoiceByAccessKey(accessKey)
	if err != nil || invoice == nil {
		c.String(http.StatusNotFound, "NF-e não encontrada")
		return
	}

	// Verify the invoice belongs to the user's farm
	dModel := departure_model.GetDepartureModel()
	departure, depErr := dModel.GetDeparture(invoice.DepartureID)
	if depErr != nil || departure.Farm != farmID {
		c.String(http.StatusForbidden, "Access denied")
		return
	}

	if invoice.Status != "authorized" {
		c.String(http.StatusBadRequest, "Somente NF-e autorizadas podem ser canceladas")
		return
	}

	pastDeadline := false
	if t, ok := invoice.AuthorizedAt.(time.Time); ok && time.Since(t) > 24*time.Hour {
		pastDeadline = true
	}

	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "nfe-cancel-modal", gin.H{
		"Invoice":      invoice,
		"PastDeadline": pastDeadline,
		"CSPNonce":     nonce.(string),
	})
}

// cancelNFe registers a cancellation event (110111) at SEFAZ for an authorized
// NF-e. On success it returns the re-rendered list row so htmx swaps it in
// place; on failure it only triggers a toast (no swap).
func cancelNFe(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	accessKey := c.Param("accessKey")
	justification := c.PostForm("justification")

	svc := nfe_service.NewNFeService()
	toast := svc.CancelInvoice(accessKey, justification, farmID)

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

	// Re-fetch the invoice so the row reflects the new 'cancelled' status
	nfeModel := nfe_model.GetNFeModel()
	invoice, err := nfeModel.GetInvoiceByAccessKey(accessKey)
	if err != nil || invoice == nil {
		c.Status(http.StatusOK)
		return
	}

	c.HTML(http.StatusOK, "nfe-list-item", invoice)
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

// parseInvoiceOverrides extracts optional per-emission override fields from the
// form submission. Empty strings yield nil pointers so the service falls back
// to the farm config defaults.
func parseInvoiceOverrides(c *gin.Context) *entity.InvoiceOverrides {
	o := &entity.InvoiceOverrides{}
	hasValue := false

	if v := strings.TrimSpace(c.PostForm("naturezaOp")); v != "" {
		o.NaturezaOp = &v
		hasValue = true
	}
	if v := strings.TrimSpace(c.PostForm("productDesc")); v != "" {
		o.ProductDesc = &v
		hasValue = true
	}
	if v := strings.TrimSpace(c.PostForm("ncm")); v != "" {
		o.NCM = &v
		hasValue = true
	}
	if v := strings.TrimSpace(c.PostForm("cest")); v != "" {
		o.CEST = &v
		hasValue = true
	}
	if v := strings.TrimSpace(c.PostForm("unit")); v != "" {
		o.Unit = &v
		hasValue = true
	}
	if v := strings.TrimSpace(c.PostForm("modFrete")); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			o.ModFrete = &i
			hasValue = true
		}
	}
	if v := strings.TrimSpace(c.PostForm("icmsCST")); v != "" {
		o.ICMSCST = &v
		hasValue = true
	}
	if v := strings.TrimSpace(c.PostForm("pisCST")); v != "" {
		o.PISCST = &v
		hasValue = true
	}
	if v := strings.TrimSpace(c.PostForm("cofinsCST")); v != "" {
		o.COFINSCST = &v
		hasValue = true
	}
	// infCpl is always parsed (even when empty) because the textarea is always
	// submitted. An empty string means the user explicitly cleared the field.
	if v, ok := c.GetPostForm("infCpl"); ok {
		o.InfCpl = &v
		hasValue = true
	}

	if !hasValue {
		return nil
	}
	return o
}

// safeFormValue returns the trimmed form value or an empty string.
func safeFormValue(c *gin.Context, key string) string {
	return strings.TrimSpace(c.PostForm(key))
}

// safePtrString returns the string pointed to by p, or an empty string if p is nil.
func safePtrString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
