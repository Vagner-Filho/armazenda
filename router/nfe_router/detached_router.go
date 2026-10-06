package nfe_router

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"armazenda/model/farm_product_model"
	"armazenda/model/nfe_model"
	"armazenda/model/person_model"
	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"
	nfe_pdf "armazenda/pkg/nfe/service"
	nfe_xml "armazenda/pkg/nfe/xml"
	"armazenda/service/nfe_service"
	"armazenda/service/user_service"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// detachedPageItem is one initial item row rendered by the detached emission
// page. Quantity and gross weight stay empty: they are inputs for each
// emission, never rascunho values.
type detachedPageItem struct {
	FarmProductID string
	ProductName   string
	NCM           string
	CEST          string
	CFOP          string
	UnitPrice     string
	Unit          string
}

// detachedPageView is the view model of the detached emission page, rendered
// either blank (farm defaults) or prefilled from a Rascunho de NF-e.
type detachedPageView struct {
	FarmID           uint32
	People           []entity_public.PersonOption
	FarmProducts     []entity_public.FarmProduct
	HasNFEFarmConfig bool
	CSPNonce         string
	TierKey          string

	// Initial item rows (at least one).
	Items []detachedPageItem

	// Recipient prefill. RecipientType is "existing" or "new".
	RecipientType         string
	SelectedPersonID      string
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

	// Rascunho context (informational; never reapplied at preview/confirm).
	HasRascunho  bool
	RascunhoID   string
	RascunhoName string

	// Warning rendered when applying a missing/cross-farm rascunho.
	Warning string

	// Operation / freight / tax defaults resolved for this render.
	DefaultNaturezaOp  string
	DefaultModFrete    int
	DefaultInfCpl      string
	UseDefaultTaxRates bool
	DefaultICMSRate    string
	DefaultPISRate     string
	DefaultCOFINSRate  string
	DefaultIBSRate     string
	DefaultCBSRate     string
	DefaultICMSCST     string
	DefaultPISCST      string
	DefaultCOFINSCST   string
	DefaultIBSCST      string
	DefaultCBSCST      string
	DefaultCClassTrib  string
}

// buildDetachedPageView assembles the emission page data. When profile is nil
// the page renders blank (one item row with the farm/default CFOP); when a
// rascunho is provided its values prefill the form. Resolution is rascunho
// value -> farm NF-e configuration -> existing system fallback, and an
// explicitly stored zero rate stays zero.
func buildDetachedPageView(farmID uint32, profile *nfe_model.DetachedProfile) detachedPageView {
	nfeModel := nfe_model.GetNFeModel()
	farmNFeConfig, _ := nfeModel.GetFarmConfig(farmID)

	pModel := person_model.GetPersonModel()
	people, _ := pModel.GetPeopleByFarm(farmID)

	farmProducts, _ := farm_product_model.GetFarmProductModel().GetFarmProductsByFarm(farmID)

	view := detachedPageView{
		FarmID:             farmID,
		People:             people,
		FarmProducts:       farmProducts,
		HasNFEFarmConfig:   farmNFeConfig != nil,
		RecipientType:      "existing",
		UseDefaultTaxRates: true,
	}

	defaultCFOP := "5101"
	if farmNFeConfig != nil {
		if farmNFeConfig.DefaultCFOP != "" {
			defaultCFOP = farmNFeConfig.DefaultCFOP
		}
		view.DefaultModFrete = farmNFeConfig.DefaultModFrete
		if farmNFeConfig.DefaultNaturezaOp != nil {
			view.DefaultNaturezaOp = *farmNFeConfig.DefaultNaturezaOp
		}
		view.DefaultICMSRate = percentDisplay(farmNFeConfig.ICMSRate)
		view.DefaultPISRate = percentDisplay(farmNFeConfig.PISRate)
		view.DefaultCOFINSRate = percentDisplay(farmNFeConfig.COFINSRate)
		view.DefaultIBSRate = percentDisplay(farmNFeConfig.IBSRate)
		view.DefaultCBSRate = percentDisplay(farmNFeConfig.CBSRate)
		view.DefaultICMSCST = safePtrString(farmNFeConfig.DefaultICMSCST)
		view.DefaultPISCST = safePtrString(farmNFeConfig.DefaultPISCST)
		view.DefaultCOFINSCST = safePtrString(farmNFeConfig.DefaultCOFINSCST)
		view.DefaultIBSCST = safePtrString(farmNFeConfig.DefaultIBSCST)
		view.DefaultCBSCST = safePtrString(farmNFeConfig.DefaultCBSCST)
		view.DefaultCClassTrib = safePtrString(farmNFeConfig.DefaultCClassTrib)

		// Derive the natureza da operação from the configured CFOP when unset.
		if view.DefaultNaturezaOp == "" {
			view.DefaultNaturezaOp = defaults.NaturezaOpForCFOP(defaultCFOP)
		}
	}

	if profile == nil {
		view.Items = []detachedPageItem{{CFOP: defaultCFOP, Unit: farmDefaultUnit(farmNFeConfig)}}
		return view
	}

	// Applied rascunho: prefill the items, recipient and operation/tax values.
	view.HasRascunho = true
	view.RascunhoID = strconv.Itoa(profile.ID)
	view.RascunhoName = profile.Name

	view.Items = make([]detachedPageItem, 0, len(profile.Items))
	for _, item := range profile.Items {
		unit := item.Unit
		if unit == "" {
			unit = farmDefaultUnit(farmNFeConfig)
		}
		cfop := item.CFOP
		if cfop == "" {
			cfop = defaultCFOP
		}
		view.Items = append(view.Items, detachedPageItem{
			FarmProductID: uint16PtrString(item.FarmProductID),
			ProductName:   item.ProductName,
			NCM:           item.NCM,
			CEST:          safePtrString(item.CEST),
			CFOP:          cfop,
			UnitPrice:     item.UnitPrice.String(),
			Unit:          unit,
		})
	}
	if len(view.Items) == 0 {
		view.Items = []detachedPageItem{{CFOP: defaultCFOP, Unit: farmDefaultUnit(farmNFeConfig)}}
	}

	if profile.RecipientID != nil {
		view.RecipientType = "existing"
		view.SelectedPersonID = strconv.FormatUint(uint64(*profile.RecipientID), 10)
	} else {
		view.RecipientType = "new"
	}

	if profile.NaturezaOp != nil && *profile.NaturezaOp != "" {
		view.DefaultNaturezaOp = *profile.NaturezaOp
	} else if view.DefaultNaturezaOp == "" {
		if common, ok := commonProfileItemCFOP(profile.Items); ok {
			view.DefaultNaturezaOp = defaults.NaturezaOpForCFOP(common)
		}
	}
	if profile.ModFrete != nil {
		view.DefaultModFrete = *profile.ModFrete
	}
	if profile.InfCpl != nil {
		view.DefaultInfCpl = *profile.InfCpl
	}

	// Tax CST / cClassTrib overrides: rascunho value wins, otherwise the farm
	// default already loaded above.
	view.DefaultICMSCST = firstNonEmpty(safePtrString(profile.ICMSCST), view.DefaultICMSCST)
	view.DefaultPISCST = firstNonEmpty(safePtrString(profile.PISCST), view.DefaultPISCST)
	view.DefaultCOFINSCST = firstNonEmpty(safePtrString(profile.COFINSCST), view.DefaultCOFINSCST)
	view.DefaultIBSCST = firstNonEmpty(safePtrString(profile.IBSCST), view.DefaultIBSCST)
	view.DefaultCBSCST = firstNonEmpty(safePtrString(profile.CBSCST), view.DefaultCBSCST)
	view.DefaultCClassTrib = firstNonEmpty(safePtrString(profile.CClassTrib), view.DefaultCClassTrib)

	// Rates: an explicit rascunho rate (including zero) is preserved; unset
	// axes keep the farm default display.
	taxRates := profile.TaxRates
	if taxRates != nil {
		explicit := false
		if taxRates.ICMSRate != nil {
			view.DefaultICMSRate = percentDisplayExact(*taxRates.ICMSRate)
			explicit = true
		}
		if taxRates.PISRate != nil {
			view.DefaultPISRate = percentDisplayExact(*taxRates.PISRate)
			explicit = true
		}
		if taxRates.COFINSRate != nil {
			view.DefaultCOFINSRate = percentDisplayExact(*taxRates.COFINSRate)
			explicit = true
		}
		if taxRates.IBSRate != nil {
			view.DefaultIBSRate = percentDisplayExact(*taxRates.IBSRate)
			explicit = true
		}
		if taxRates.CBSRate != nil {
			view.DefaultCBSRate = percentDisplayExact(*taxRates.CBSRate)
			explicit = true
		}
		view.UseDefaultTaxRates = !explicit
	}

	return view
}

// getDetachedNFePage renders the full page for detached NF-e emission.
func getDetachedNFePage(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	view := buildDetachedPageView(farmID, nil)
	nonce, _ := c.Get("csp_nonce")
	view.CSPNonce = nonce.(string)
	view.TierKey = user_service.GetTierKeyFromContext(c)

	c.HTML(http.StatusOK, "nfe-emit.html", view)
}

// getDetachedNFePageFromProfile renders the detached emission page prefilled
// from a farm-scoped Rascunho de NF-e. A missing or cross-farm rascunho
// renders the blank page with a "Rascunho não encontrado" warning, never a
// server error.
func getDetachedNFePageFromProfile(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	profileID, err := strconv.Atoi(c.Param("rascunhoId"))
	if err != nil || profileID < 1 {
		view := buildDetachedPageView(farmID, nil)
		view.Warning = "Rascunho não encontrado"
		renderDetachedPage(c, view)
		return
	}

	svc := nfe_service.NewNFeService()
	profile, toast := svc.GetDetachedProfile(farmID, profileID)
	if isFailureToast(toast) {
		view := buildDetachedPageView(farmID, nil)
		view.Warning = toast.Message
		renderDetachedPage(c, view)
		return
	}

	renderDetachedPage(c, buildDetachedPageView(farmID, profile))
}

// renderDetachedPage attaches the request-scoped template data and renders the
// emission page.
func renderDetachedPage(c *gin.Context, view detachedPageView) {
	nonce, _ := c.Get("csp_nonce")
	view.CSPNonce = nonce.(string)
	view.TierKey = user_service.GetTierKeyFromContext(c)
	c.HTML(http.StatusOK, "nfe-emit.html", view)
}

// farmDefaultUnit returns the farm default unit, falling back to KG.
func farmDefaultUnit(farmNFeConfig *entity_public.FarmConfig) string {
	if farmNFeConfig != nil && farmNFeConfig.DefaultUnit != "" {
		return farmNFeConfig.DefaultUnit
	}
	return "KG"
}

// firstNonEmpty returns value when non-empty, otherwise fallback.
func firstNonEmpty(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

// percentDisplayExact formats a rate (0.17) as a fixed two-decimal percentage
// string ("17.00"), preserving an explicit zero as "0.00".
func percentDisplayExact(rate decimal.Decimal) string {
	return rate.Mul(decimal.NewFromInt(100)).StringFixed(2)
}

// commonProfileItemCFOP returns the shared CFOP when every rascunho item uses
// the same code.
func commonProfileItemCFOP(items []nfe_model.DetachedProfileItem) (string, bool) {
	if len(items) == 0 || items[0].CFOP == "" {
		return "", false
	}
	first := items[0].CFOP
	for _, item := range items[1:] {
		if item.CFOP != first {
			return "", false
		}
	}
	return first, true
}

// detachedFormData carries every field parsed from the detached NF-e form.
// The preview and the emission handlers both build their
// nfe_service.DetachedInvoiceInput from it so the two paths cannot drift.
type detachedFormData struct {
	Recipient nfe_service.DetachedRecipient
	Items     []nfe_service.DetachedItemInput
	TaxRates  entity.TaxRates
	Overrides *entity.InvoiceOverrides
	VehicleID *uint16
	ProfileID string // informational only: the rascunho the form started from
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
		TaxRates:  userRates,
		Overrides: parseInvoiceOverrides(c),
		VehicleID: vehicleID,
		ProfileID: c.PostForm("rascunhoId"),
	}, entity_public.Toast{}
}

// detachedInputFromForm maps the parsed form into the service input.
func detachedInputFromForm(farmID uint32, form detachedFormData) nfe_service.DetachedInvoiceInput {
	return nfe_service.DetachedInvoiceInput{
		FarmID:    farmID,
		Recipient: form.Recipient,
		VehicleID: form.VehicleID,
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
	CFOP          string
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

	ProfileID  string
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

// detachedPreviewItems maps parsed items to their hidden-field representation.
// It is a pure function so the preview-to-confirm round trip (including the
// per-item CFOP and exact price strings) is unit-testable.
func detachedPreviewItems(parsed []nfe_service.DetachedItemInput) []detachedPreviewItem {
	items := make([]detachedPreviewItem, 0, len(parsed))
	for i, item := range parsed {
		items = append(items, detachedPreviewItem{
			Index:         i,
			FarmProductID: uint16PtrString(item.FarmProductID),
			ProductName:   item.ProductName,
			NCM:           item.NCM,
			CEST:          safePtrString(item.CEST),
			CFOP:          item.CFOP,
			Quantity:      item.Quantity.String(),
			GrossWeight:   item.GrossWeight.String(),
			UnitPrice:     item.UnitPrice.String(),
			Unit:          item.Unit,
		})
	}
	return items
}

// buildDetachedPreviewViewData assembles the view model for the
// nfe-detached-preview fragment. Everything except the display-only
// serie/environment comes from the parsed form, so the hidden fields
// reproduce the exact input the preview was generated from.
func buildDetachedPreviewViewData(farmID uint32, form detachedFormData, pdfBytes []byte, nonce string) detachedPreviewView {
	items := detachedPreviewItems(form.Items)
	var totalValue decimal.Decimal
	for _, item := range form.Items {
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
		ProfileID:             form.ProfileID,
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

		// Parse the per-item CFOP (required: exactly four ASCII digits)
		cfop := strings.TrimSpace(c.PostForm(prefix + "cfop"))
		if !nfe_service.IsFourDigitCFOP(cfop) {
			return nil, entity_public.GetWarningToast(
				"CFOP inválido",
				fmt.Sprintf("Item %d: informe um CFOP com exatamente 4 dígitos", i+1))
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
			CFOP:          cfop,
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

// danfeStatusAllowed reports whether a DANFE may be rendered for the given
// invoice status. Per MOC Anexo II only authorized invoices get a DANFE,
// plus cancelled ones rendered with the "NF-e CANCELADA" banner.
func danfeStatusAllowed(status string) bool {
	return status == "authorized" || status == "cancelled"
}

// detachedDANFEXML returns the XML to render the DANFE from, preferring the
// authorized document. ok is false when neither XML is stored.
func detachedDANFEXML(invoice *nfe_model.DetachedInvoice) (string, bool) {
	if invoice.XMLAuthorized != nil && *invoice.XMLAuthorized != "" {
		return *invoice.XMLAuthorized, true
	}
	if invoice.XMLSigned != nil && *invoice.XMLSigned != "" {
		return *invoice.XMLSigned, true
	}
	return "", false
}

// downloadDetachedNFeDANFE downloads the DANFE of a detached invoice. It
// renders the fiscal DANFE from the stored XML (what SEFAZ authorized);
// cancelled invoices render with the "NF-e CANCELADA" banner per MOC Anexo II.
func downloadDetachedNFeDANFE(c *gin.Context) {
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

	// DANFE is available for authorized invoices, and for cancelled invoices
	// rendered with an "NF-e CANCELADA" banner per MOC Anexo II.
	if !danfeStatusAllowed(invoice.Status) {
		c.String(http.StatusForbidden, "DANFE somente disponivel para NF-e autorizada pela SEFAZ")
		return
	}

	xmlContent, ok := detachedDANFEXML(invoice)
	if !ok {
		c.String(http.StatusNotFound, "XML não disponível")
		return
	}

	data, parseErr := nfe_xml.ParseDANFEData(xmlContent)
	if parseErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("downloadDetachedNFeDANFE parse error: %v", parseErr.Error()))
		c.String(http.StatusInternalServerError, "Falha ao processar os dados da NF-e")
		return
	}

	// Fallback: if the XML doesn't contain <protNFe> (e.g., invoices
	// authorized synchronously, where the <nfeProc> wrapper is not stored),
	// populate the protocol from the database column so Campo 2 on the DANFE
	// is filled.
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
		model_error.GetLoggerModel().Log(fmt.Sprintf("downloadDetachedNFeDANFE generate error: %v", genErr.Error()))
		c.String(http.StatusInternalServerError, "Falha ao gerar a DANFE")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=danfe_%s.pdf", accessKey))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
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
	router.GET("/emitir/rascunho/:rascunhoId", getDetachedNFePageFromProfile)
	router.POST("/emitir/preview", previewDetachedNFe)
	router.POST("/emitir/build", buildDetachedNFe)
	router.GET("/avulsa/list", getDetachedNFeList)
	router.GET("/avulsa/download/xml/:accessKey", downloadDetachedNFeXML)
	router.GET("/avulsa/download/danfe/:accessKey", downloadDetachedNFeDANFE)
	router.POST("/avulsa/cancel/:accessKey", cancelDetachedNFe)

	// Rascunhos de NF-e management + save-from-form
	UseDetachedProfileRoutes(router)
}
