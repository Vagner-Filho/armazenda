package nfe_router

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	entity_public "armazenda/entity/public"
	"armazenda/model/farm_product_model"
	"armazenda/model/nfe_model"
	"armazenda/model/person_model"
	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"
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

	// Get the farm's own products for the per-item selector
	farmProducts, _ := farm_product_model.GetFarmProductModel().GetFarmProductsByFarm(farmID)

	// Resolve the farm-config defaults that pre-fill the tax and transport
	// inputs. Missing config is non-fatal: the form renders with empty
	// defaults and the "no config" banner blocks emission.
	var defaultICMS, defaultPIS, defaultCOFINS, defaultIBS, defaultCBS decimal.Decimal
	var defaultNaturezaOp, defaultCEST, defaultICMSCST, defaultPISCST, defaultCOFINSCST *string
	var defaultCBSCST, defaultIBSCST, defaultCClassTrib *string
	var defaultModFrete int
	if farmNFeConfig != nil {
		defaultICMS = farmNFeConfig.ICMSRate
		defaultPIS = farmNFeConfig.PISRate
		defaultCOFINS = farmNFeConfig.COFINSRate
		defaultIBS = farmNFeConfig.IBSRate
		defaultCBS = farmNFeConfig.CBSRate
		defaultNaturezaOp = farmNFeConfig.DefaultNaturezaOp
		defaultCEST = farmNFeConfig.DefaultCEST
		defaultModFrete = farmNFeConfig.DefaultModFrete
		defaultICMSCST = farmNFeConfig.DefaultICMSCST
		defaultPISCST = farmNFeConfig.DefaultPISCST
		defaultCOFINSCST = farmNFeConfig.DefaultCOFINSCST
		defaultCBSCST = farmNFeConfig.DefaultCBSCST
		defaultIBSCST = farmNFeConfig.DefaultIBSCST
		defaultCClassTrib = farmNFeConfig.DefaultCClassTrib

		// Derive the natureza da operação from the configured CFOP when unset
		if defaultNaturezaOp == nil || *defaultNaturezaOp == "" {
			derived := defaults.NaturezaOpForCFOP(farmNFeConfig.DefaultCFOP)
			defaultNaturezaOp = &derived
		}
	}

	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "nfe-emit.html", gin.H{
		"FarmID":            farmID,
		"People":            people,
		"FarmProducts":      farmProducts,
		"HasNFEFarmConfig":  farmNFeConfig != nil,
		"DefaultCFOP":       "5101",
		"DefaultUnit":       "KG",
		"DefaultICMSRate":   percentDisplay(defaultICMS),
		"DefaultPISRate":    percentDisplay(defaultPIS),
		"DefaultCOFINSRate": percentDisplay(defaultCOFINS),
		"DefaultIBSRate":    percentDisplay(defaultIBS),
		"DefaultCBSRate":    percentDisplay(defaultCBS),
		"DefaultNaturezaOp": safePtrString(defaultNaturezaOp),
		"DefaultCEST":       safePtrString(defaultCEST),
		"DefaultModFrete":   defaultModFrete,
		"DefaultICMSCST":    safePtrString(defaultICMSCST),
		"DefaultPISCST":     safePtrString(defaultPISCST),
		"DefaultCOFINSCST":  safePtrString(defaultCOFINSCST),
		"DefaultIBSCST":     safePtrString(defaultIBSCST),
		"DefaultCBSCST":     safePtrString(defaultCBSCST),
		"DefaultCClassTrib": safePtrString(defaultCClassTrib),
		"CSPNonce":          nonce.(string),
		"TierKey":           user_service.GetTierKeyFromContext(c),
	})
}

// detachedFormData carries every field parsed from the detached NF-e form.
// The preview and the emission handlers both build their
// nfe_service.DetachedInvoiceInput from it so the two paths cannot drift.
type detachedFormData struct {
	Recipient nfe_service.DetachedRecipient
	Items     []nfe_service.DetachedItemInput
	CFOP      string
	TaxRates  entity.TaxRates
	Overrides *entity.InvoiceOverrides
	VehicleID *uint16
}

// isFailureToast reports whether a toast represents a failure the request must
// not proceed past.
func isFailureToast(toast entity_public.Toast) bool {
	return toast.Type == entity_public.ErrorToast || toast.Type == entity_public.WarningToast
}

// parseDetachedForm parses every field the detached emission form submits.
// Preview and emission consume the same parsed data, so the previewed values
// are exactly what the confirm step re-submits.
func parseDetachedForm(c *gin.Context) (detachedFormData, entity_public.Toast) {
	recipient, toast := parseDetachedRecipient(c)
	if isFailureToast(toast) {
		return detachedFormData{}, toast
	}

	items, itemsToast := parseDetachedItems(c)
	if isFailureToast(itemsToast) {
		return detachedFormData{}, itemsToast
	}

	cfop := c.PostForm("cfop")
	if cfop == "" {
		cfop = "5101"
	}

	userRates, rateErr := parseUserTaxRates(c)
	if rateErr != nil {
		return detachedFormData{}, entity_public.GetWarningToast("Falha ao identificar as alíquotas", rateErr.Error())
	}

	// Parse vehicle (optional)
	var vehicleID *uint16
	if vehicleStr := c.PostForm("vehicleId"); vehicleStr != "" {
		if vid, err := strconv.ParseUint(vehicleStr, 10, 16); err == nil {
			v := uint16(vid)
			vehicleID = &v
		}
	}

	return detachedFormData{
		Recipient: recipient,
		Items:     items,
		CFOP:      cfop,
		TaxRates:  userRates,
		Overrides: parseInvoiceOverrides(c),
		VehicleID: vehicleID,
	}, entity_public.Toast{}
}

// detachedInputFromForm maps the parsed form into the service input.
func detachedInputFromForm(farmID uint32, form detachedFormData) nfe_service.DetachedInvoiceInput {
	return nfe_service.DetachedInvoiceInput{
		FarmID:    farmID,
		Recipient: form.Recipient,
		VehicleID: form.VehicleID,
		CFOP:      form.CFOP,
		TaxRates:  form.TaxRates,
		Overrides: form.Overrides,
		Items:     form.Items,
	}
}

// buildDetachedNFe handles the form submission to emit a detached NF-e
func buildDetachedNFe(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	form, toast := parseDetachedForm(c)
	if isFailureToast(toast) {
		c.Header("HX-Trigger", toast.ToJsonStr())
		c.Status(http.StatusBadRequest)
		return
	}

	svc := nfe_service.NewNFeService()
	result, toast := svc.BuildDetachedInvoice(detachedInputFromForm(farmID, form))

	if isFailureToast(toast) {
		c.Header("HX-Trigger", toast.ToJsonStr())
		if toast.Type == entity_public.WarningToast {
			c.Status(http.StatusBadRequest)
		} else {
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	// Success: the toast carries the SEFAZ status and the fragment offers the
	// signed XML for download (served by the existing access-key route).
	c.Header("HX-Trigger", toast.ToJsonStr())
	c.HTML(http.StatusOK, "nfe-detached-result", gin.H{
		"AccessKey": result.AccessKey,
	})
}

// detachedPreviewItem is the hidden-field representation of one parsed item.
type detachedPreviewItem struct {
	Index         int
	FarmProductID string
	ProductName   string
	NCM           string
	CEST          string
	Quantity      string
	GrossWeight   string
	UnitPrice     string
	Unit          string
}

// detachedPreviewView is the view model for the nfe-detached-preview
// fragment: the preview PDF plus every parsed field rendered as a hidden
// input so the confirm step is stateless.
type detachedPreviewView struct {
	PDFBase64   string
	Serie       int
	Environment int
	CSPNonce    string

	// Summary (display only)
	RecipientLabel string
	TotalValue     decimal.Decimal
	ItemCount      int

	// Recipient: PersonID when an existing person was selected, otherwise
	// the inline fields.
	PersonID              string
	RecipientName         string
	RecipientDocument     string
	RecipientIE           string
	RecipientStreet       string
	RecipientNumber       string
	RecipientNeighborhood string
	RecipientCity         string
	RecipientState        string
	RecipientCEP          string
	RecipientPhone        string
	RecipientEmail        string

	Items []detachedPreviewItem

	CFOP       string
	NaturezaOp string
	InfCpl     string
	ModFrete   string

	// Rates (percentage strings; empty means "use the farm config default")
	ICMSRate   string
	PISRate    string
	COFINSRate string
	IBSRate    string
	CBSRate    string

	// CST / cClassTrib overrides (empty means "not provided")
	ICMSCST    string
	PISCST     string
	COFINSCST  string
	IBSCST     string
	CBSCST     string
	CClassTrib string
}

// buildDetachedPreviewViewData assembles the view model for the
// nfe-detached-preview fragment. Everything except the display-only
// serie/environment comes from the parsed form, so the hidden fields
// reproduce the exact input the preview was generated from.
func buildDetachedPreviewViewData(farmID uint32, form detachedFormData, pdfBytes []byte, nonce string) detachedPreviewView {
	items := make([]detachedPreviewItem, 0, len(form.Items))
	var totalValue decimal.Decimal
	for i, item := range form.Items {
		items = append(items, detachedPreviewItem{
			Index:         i,
			FarmProductID: uint16PtrString(item.FarmProductID),
			ProductName:   item.ProductName,
			NCM:           item.NCM,
			CEST:          safePtrString(item.CEST),
			Quantity:      item.Quantity.String(),
			GrossWeight:   item.GrossWeight.String(),
			UnitPrice:     item.UnitPrice.String(),
			Unit:          item.Unit,
		})
		totalValue = totalValue.Add(item.Quantity.Mul(item.UnitPrice))
	}

	// Serie/environment are display-only context; a missing config is
	// non-fatal here because the preview service would have rejected the
	// request already.
	serie := 0
	environment := 1
	if cfg, err := nfe_model.GetNFeModel().GetFarmConfig(farmID); err == nil && cfg != nil {
		serie = cfg.Serie
		environment = cfg.Environment
	}

	recipientLabel := form.Recipient.Name
	if form.Recipient.PersonID != nil {
		recipientLabel = "Destinatário selecionado"
		if person, err := person_model.GetPersonModel().GetFullPersonById(*form.Recipient.PersonID); err == nil {
			recipientLabel = person.Name
		}
	}

	view := detachedPreviewView{
		PDFBase64:             base64.StdEncoding.EncodeToString(pdfBytes),
		Serie:                 serie,
		Environment:           environment,
		CSPNonce:              nonce,
		RecipientLabel:        recipientLabel,
		TotalValue:            totalValue,
		ItemCount:             len(items),
		PersonID:              uint32PtrString(form.Recipient.PersonID),
		RecipientName:         form.Recipient.Name,
		RecipientDocument:     form.Recipient.Document,
		RecipientIE:           safePtrString(form.Recipient.IE),
		RecipientStreet:       safePtrString(form.Recipient.Street),
		RecipientNumber:       safePtrString(form.Recipient.Number),
		RecipientNeighborhood: safePtrString(form.Recipient.Neighborhood),
		RecipientCity:         safePtrString(form.Recipient.City),
		RecipientState:        safePtrString(form.Recipient.State),
		RecipientCEP:          safePtrString(form.Recipient.CEP),
		RecipientPhone:        safePtrString(form.Recipient.PhoneNumber),
		RecipientEmail:        safePtrString(form.Recipient.Email),
		Items:                 items,
		CFOP:                  form.CFOP,
		ICMSRate:              rateDisplayString(form.TaxRates.ICMSRate),
		PISRate:               rateDisplayString(form.TaxRates.PISRate),
		COFINSRate:            rateDisplayString(form.TaxRates.COFINSRate),
		IBSRate:               rateDisplayString(form.TaxRates.IBSRate),
		CBSRate:               rateDisplayString(form.TaxRates.CBSRate),
	}

	if o := form.Overrides; o != nil {
		view.NaturezaOp = safePtrString(o.NaturezaOp)
		view.InfCpl = safePtrString(o.InfCpl)
		if o.ModFrete != nil {
			view.ModFrete = strconv.Itoa(*o.ModFrete)
		}
		view.ICMSCST = safePtrString(o.ICMSCST)
		view.PISCST = safePtrString(o.PISCST)
		view.COFINSCST = safePtrString(o.COFINSCST)
		view.IBSCST = safePtrString(o.IBSCST)
		view.CBSCST = safePtrString(o.CBSCST)
		view.CClassTrib = safePtrString(o.CClassTrib)
	}

	return view
}

// previewDetachedNFe renders the non-fiscal DANFE preview for a detached
// NF-e. It creates no person, allocates no number and persists nothing; the
// fragment carries all parsed fields as hidden inputs for the confirm step.
func previewDetachedNFe(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	form, toast := parseDetachedForm(c)
	if isFailureToast(toast) {
		c.Header("HX-Trigger", toast.ToJsonStr())
		c.Status(http.StatusBadRequest)
		return
	}

	svc := nfe_service.NewNFeService()
	pdfBytes, toast := svc.GenerateDetachedPreviewDANFE(detachedInputFromForm(farmID, form))
	if isFailureToast(toast) {
		c.Header("HX-Trigger", toast.ToJsonStr())
		if toast.Type == entity_public.WarningToast {
			c.Status(http.StatusBadRequest)
		} else {
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "nfe-detached-preview", buildDetachedPreviewViewData(farmID, form, pdfBytes, nonce.(string)))
}

// uint16PtrString formats an optional uint16 identifier for a hidden input.
func uint16PtrString(p *uint16) string {
	if p == nil {
		return ""
	}
	return strconv.FormatUint(uint64(*p), 10)
}

// uint32PtrString formats an optional uint32 identifier for a hidden input.
func uint32PtrString(p *uint32) string {
	if p == nil {
		return ""
	}
	return strconv.FormatUint(uint64(*p), 10)
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
	router.POST("/emitir/preview", previewDetachedNFe)
	router.POST("/emitir/build", buildDetachedNFe)
	router.GET("/avulsa/list", getDetachedNFeList)
	router.GET("/avulsa/download/xml/:accessKey", downloadDetachedNFeXML)
	router.GET("/avulsa/download/danfe/:accessKey", downloadDetachedNFeDANFE)
	router.POST("/avulsa/cancel/:accessKey", cancelDetachedNFe)
}
