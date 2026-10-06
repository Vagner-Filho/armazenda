package nfe_router

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	entity_public "armazenda/entity/public"
	"armazenda/model/farm_product_model"
	"armazenda/model/nfe_model"
	"armazenda/model/person_model"
	"armazenda/service/nfe_service"
	"armazenda/service/user_service"
	cfop_view "armazenda/view/cfop"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// getDetachedProfilesPage renders the "Rascunhos de NF-e" management page.
func getDetachedProfilesPage(c *gin.Context) {
	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "nfe-rascunhos.html", gin.H{
		"CSPNonce": nonce.(string),
		"TierKey":  user_service.GetTierKeyFromContext(c),
	})
}

// getDetachedProfileList renders the farm's rascunho table body.
func getDetachedProfileList(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	svc := nfe_service.NewNFeService()
	profiles, toast := svc.ListDetachedProfiles(farmID)
	if toast.Type == entity_public.ErrorToast {
		c.Header("HX-Trigger", toast.ToJsonStr())
		c.Status(http.StatusInternalServerError)
		return
	}

	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "nfe-rascunho-table", gin.H{
		"Profiles": profiles,
		"CSPNonce": nonce.(string),
	})
}

// detachedProfileFormRow is the editor's display representation of one item
// default. Strings keep a brand-new row blank instead of showing zero values.
type detachedProfileFormRow struct {
	FarmProductID string
	ProductName   string
	NCM           string
	CEST          string
	CFOP          string
	Unit          string
	UnitPrice     string
}

// detachedProfileFormRows maps a profile (or a blank draft) to editor rows.
func detachedProfileFormRows(profile *nfe_model.DetachedProfile) []detachedProfileFormRow {
	rows := make([]detachedProfileFormRow, 0, 1)
	if profile != nil {
		for _, item := range profile.Items {
			rows = append(rows, detachedProfileFormRow{
				FarmProductID: uint16PtrString(item.FarmProductID),
				ProductName:   item.ProductName,
				NCM:           item.NCM,
				CEST:          safePtrString(item.CEST),
				CFOP:          item.CFOP,
				Unit:          item.Unit,
				UnitPrice:     item.UnitPrice.String(),
			})
		}
	}
	if len(rows) == 0 {
		rows = append(rows, detachedProfileFormRow{Unit: "KG"})
	}
	return rows
}

// detachedProfileFormView builds the editor dialog view data. Explicit rates
// (including zero) are displayed exactly; unset axes stay blank so they keep
// inheriting the farm configuration.
func detachedProfileFormView(profile *nfe_model.DetachedProfile, people []entity_public.PersonOption, farmProducts []entity_public.FarmProduct, cfopOptions cfop_view.CfopOptions, nonce string) gin.H {
	view := gin.H{
		"Profile":            profile,
		"People":             people,
		"FarmProducts":       farmProducts,
		"CfopOptions":        cfopOptions,
		"DefaultCFOP":        cfopOptions.Default,
		"CSPNonce":           nonce,
		"RecipientSelected":  "",
		"Rows":               detachedProfileFormRows(profile),
		"ICMSRate":           "",
		"PISRate":            "",
		"COFINSRate":         "",
		"IBSRate":            "",
		"CBSRate":            "",
		"UseDefaultTaxRates": true,
		"ModFreteSelected":   0,
	}
	if profile == nil {
		return view
	}
	if profile.RecipientID != nil {
		view["RecipientSelected"] = strconv.FormatUint(uint64(*profile.RecipientID), 10)
	}
	if profile.ModFrete != nil {
		view["ModFreteSelected"] = *profile.ModFrete
	}
	if profile.TaxRates != nil {
		explicit := false
		view["ICMSRate"] = profileRateDisplay(profile.TaxRates.ICMSRate)
		view["PISRate"] = profileRateDisplay(profile.TaxRates.PISRate)
		view["COFINSRate"] = profileRateDisplay(profile.TaxRates.COFINSRate)
		view["IBSRate"] = profileRateDisplay(profile.TaxRates.IBSRate)
		view["CBSRate"] = profileRateDisplay(profile.TaxRates.CBSRate)
		for _, rate := range []string{
			view["ICMSRate"].(string), view["PISRate"].(string), view["COFINSRate"].(string),
			view["IBSRate"].(string), view["CBSRate"].(string),
		} {
			if rate != "" {
				explicit = true
				break
			}
		}
		view["UseDefaultTaxRates"] = !explicit
	}
	return view
}

// profileRateDisplay formats an optional profile rate for a form input. Nil
// stays empty ("inherit"), while an explicit zero is shown as "0.00".
func profileRateDisplay(rate *decimal.Decimal) string {
	if rate == nil {
		return ""
	}
	return percentDisplayExact(*rate)
}

// getDetachedProfileForm renders the "Novo Rascunho" editor dialog.
func getDetachedProfileForm(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	pModel := person_model.GetPersonModel()
	people, _ := pModel.GetPeopleByFarm(farmID)
	farmProducts, _ := farm_product_model.GetFarmProductModel().GetFarmProductsByFarm(farmID)

	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "nfe-rascunho-form", detachedProfileFormView(nil, people, farmProducts, loadCfopOptions(c, farmID), nonce.(string)))
}

// getFilledDetachedProfileForm renders the editor dialog for an existing
// farm-scoped rascunho.
func getFilledDetachedProfileForm(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		toast := entity_public.GetWarningToast("Rascunho não encontrado", "")
		c.Header("HX-Trigger", toast.ToJsonStr())
		c.Status(http.StatusNotFound)
		return
	}

	svc := nfe_service.NewNFeService()
	profile, toast := svc.GetDetachedProfile(farmID, id)
	if isFailureToast(toast) {
		c.Header("HX-Trigger", toast.ToJsonStr())
		if toast.Type == entity_public.WarningToast {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	pModel := person_model.GetPersonModel()
	people, _ := pModel.GetPeopleByFarm(farmID)
	farmProducts, _ := farm_product_model.GetFarmProductModel().GetFarmProductsByFarm(farmID)

	nonce, _ := c.Get("csp_nonce")
	c.HTML(http.StatusOK, "nfe-rascunho-form", detachedProfileFormView(profile, people, farmProducts, loadCfopOptions(c, farmID), nonce.(string)))
}

// parseDetachedProfileItems parses the editor item rows. Quantity and gross
// weight are intentionally not read: they are emission inputs, not rascunho
// values.
func parseDetachedProfileItems(c *gin.Context) ([]nfe_service.DetachedProfileItemInput, entity_public.Toast) {
	itemCountStr := c.PostForm("itemCount")
	itemCount, err := strconv.Atoi(itemCountStr)
	if err != nil || itemCount < 1 {
		return nil, entity_public.GetWarningToast("Nenhum item informado", "Adicione pelo menos um item ao rascunho")
	}

	var items []nfe_service.DetachedProfileItemInput
	for i := 0; i < itemCount; i++ {
		prefix := fmt.Sprintf("items[%d].", i)

		var farmProductID *uint16
		if fpIDStr := c.PostForm(prefix + "farmProductId"); fpIDStr != "" {
			if fpID, parseErr := strconv.ParseUint(fpIDStr, 10, 16); parseErr == nil {
				v := uint16(fpID)
				farmProductID = &v
			}
		}

		cfop := strings.TrimSpace(c.PostForm(prefix + "cfop"))
		if !nfe_service.IsFourDigitCFOP(cfop) {
			return nil, entity_public.GetWarningToast(
				"CFOP inválido",
				fmt.Sprintf("Item %d: informe um CFOP com exatamente 4 dígitos", i+1))
		}

		unitPriceStr := strings.TrimSpace(c.PostForm(prefix + "unitPrice"))
		unitPrice, parseErr := decimal.NewFromString(unitPriceStr)
		if parseErr != nil || unitPrice.IsNegative() {
			return nil, entity_public.GetWarningToast(
				"Preço unitário inválido",
				fmt.Sprintf("Item %d: informe um preço unitário válido", i+1))
		}

		var cest *string
		if v := strings.TrimSpace(c.PostForm(prefix + "cest")); v != "" {
			cest = &v
		}

		unit := strings.TrimSpace(c.PostForm(prefix + "unit"))
		if unit == "" {
			unit = "KG"
		}

		items = append(items, nfe_service.DetachedProfileItemInput{
			FarmProductID: farmProductID,
			ProductName:   strings.TrimSpace(c.PostForm(prefix + "productName")),
			NCM:           strings.TrimSpace(c.PostForm(prefix + "ncm")),
			CEST:          cest,
			CFOP:          cfop,
			Unit:          unit,
			UnitPrice:     unitPrice,
		})
	}

	return items, entity_public.Toast{}
}

// parseDetachedProfileRecipient parses the optional recipient of a rascunho.
// Inline data is allowed because the save-from-form action explicitly creates
// the contact; the editor dialog submits a registered person id only.
func parseDetachedProfileRecipient(c *gin.Context) (nfe_service.DetachedRecipient, entity_public.Toast) {
	if personIDStr := c.PostForm("recipientPersonId"); personIDStr != "" {
		if pid, err := strconv.ParseUint(personIDStr, 10, 32); err == nil {
			personID := uint32(pid)
			return nfe_service.DetachedRecipient{PersonID: &personID}, entity_public.Toast{}
		}
	}

	name := strings.TrimSpace(c.PostForm("recipientName"))
	document := strings.TrimSpace(c.PostForm("recipientDocument"))
	if name == "" && document == "" {
		return nfe_service.DetachedRecipient{}, entity_public.Toast{}
	}
	if name == "" || document == "" {
		return nfe_service.DetachedRecipient{}, entity_public.GetWarningToast(
			"Destinatário incompleto",
			"Informe o nome e o documento do destinatário")
	}

	recipient := nfe_service.DetachedRecipient{Name: name, Document: document}
	if v := strings.TrimSpace(c.PostForm("recipientIE")); v != "" {
		recipient.IE = &v
	}
	if v := strings.TrimSpace(c.PostForm("recipientStreet")); v != "" {
		recipient.Street = &v
	}
	if v := strings.TrimSpace(c.PostForm("recipientNumber")); v != "" {
		recipient.Number = &v
	}
	if v := strings.TrimSpace(c.PostForm("recipientNeighborhood")); v != "" {
		recipient.Neighborhood = &v
	}
	if v := strings.TrimSpace(c.PostForm("recipientCity")); v != "" {
		recipient.City = &v
	}
	if v := strings.TrimSpace(c.PostForm("recipientState")); v != "" {
		recipient.State = &v
	}
	if v := strings.TrimSpace(c.PostForm("recipientCEP")); v != "" {
		recipient.CEP = &v
	}
	if v := strings.TrimSpace(c.PostForm("recipientPhone")); v != "" {
		recipient.PhoneNumber = &v
	}
	if v := strings.TrimSpace(c.PostForm("recipientEmail")); v != "" {
		recipient.Email = &v
	}
	return recipient, entity_public.Toast{}
}

// parseDetachedProfileForm parses the rascunho editor submission.
func parseDetachedProfileForm(c *gin.Context) (nfe_service.DetachedProfileInput, entity_public.Toast) {
	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		return nfe_service.DetachedProfileInput{}, entity_public.GetWarningToast("Nome obrigatório", "Informe um nome para o rascunho")
	}

	recipient, toast := parseDetachedProfileRecipient(c)
	if isFailureToast(toast) {
		return nfe_service.DetachedProfileInput{}, toast
	}

	items, itemsToast := parseDetachedProfileItems(c)
	if isFailureToast(itemsToast) {
		return nfe_service.DetachedProfileInput{}, itemsToast
	}

	rates, rateErr := parseUserTaxRates(c)
	if rateErr != nil {
		return nfe_service.DetachedProfileInput{}, entity_public.GetWarningToast("Falha ao identificar as alíquotas", rateErr.Error())
	}

	var modFrete *int
	if v := strings.TrimSpace(c.PostForm("modFrete")); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			modFrete = &i
		}
	}

	var naturezaOp, infCpl *string
	if v := strings.TrimSpace(c.PostForm("naturezaOp")); v != "" {
		naturezaOp = &v
	}
	if v, ok := c.GetPostForm("infCpl"); ok {
		infCpl = &v
	}

	return nfe_service.DetachedProfileInput{
		FarmID:     0, // set by the caller from the authenticated session
		Name:       name,
		Recipient:  recipient,
		NaturezaOp: naturezaOp,
		ModFrete:   modFrete,
		InfCpl:     infCpl,
		Items:      items,
		TaxRates:   rates,
		Overrides:  parseInvoiceOverrides(c),
	}, entity_public.Toast{}
}

// addDetachedProfile creates a rascunho from the editor dialog.
func addDetachedProfile(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	input, toast := parseDetachedProfileForm(c)
	if isFailureToast(toast) {
		c.Header("HX-Trigger", toast.ToJsonStr())
		c.Status(http.StatusBadRequest)
		return
	}
	input.FarmID = farmID

	svc := nfe_service.NewNFeService()
	profile, saveToast := svc.SaveDetachedProfile(input)
	if isFailureToast(saveToast) {
		c.Header("HX-Trigger", saveToast.ToJsonStr())
		if saveToast.Type == entity_public.WarningToast {
			c.Status(http.StatusBadRequest)
		} else {
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	c.Header("HX-Trigger", saveToast.ToJsonStr())
	c.HTML(http.StatusCreated, "nfe-rascunho-list-item", profile)
}

// putDetachedProfile updates a rascunho from the editor dialog.
func putDetachedProfile(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		toast := entity_public.GetWarningToast("Rascunho não encontrado", "")
		c.Header("HX-Trigger", toast.ToJsonStr())
		c.Status(http.StatusBadRequest)
		return
	}

	input, toast := parseDetachedProfileForm(c)
	if isFailureToast(toast) {
		c.Header("HX-Trigger", toast.ToJsonStr())
		c.Status(http.StatusBadRequest)
		return
	}
	input.FarmID = farmID
	input.ID = id

	svc := nfe_service.NewNFeService()
	profile, saveToast := svc.SaveDetachedProfile(input)
	if isFailureToast(saveToast) {
		c.Header("HX-Trigger", saveToast.ToJsonStr())
		if saveToast.Type == entity_public.WarningToast {
			c.Status(http.StatusBadRequest)
		} else {
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	c.Header("HX-Trigger", saveToast.ToJsonStr())
	c.HTML(http.StatusOK, "nfe-rascunho-list-item", profile)
}

// deleteDetachedProfile hard-deletes a rascunho; issued invoices keep their
// own snapshots and are never touched.
func deleteDetachedProfile(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		toast := entity_public.GetWarningToast("Rascunho não encontrado", "")
		c.Header("HX-Trigger", toast.ToJsonStr())
		c.Status(http.StatusBadRequest)
		return
	}

	svc := nfe_service.NewNFeService()
	toast := svc.DeleteDetachedProfile(farmID, id)
	c.Header("HX-Trigger", toast.ToJsonStr())
	if toast.Type == entity_public.ErrorToast {
		c.Status(http.StatusInternalServerError)
		return
	}
	if toast.Type == entity_public.WarningToast {
		c.Status(http.StatusBadRequest)
		return
	}
	c.Status(http.StatusOK)
}

// saveDetachedProfileFromEmissionForm handles "Salvar como rascunho" on the
// detached emission form. It parses the populated form (quantity/gross weight
// are ignored), creates or explicitly updates the rascunho, and performs no
// emission side effect. An inline recipient is saved as a farm contact only
// because the user explicitly pressed this action.
func saveDetachedProfileFromEmissionForm(c *gin.Context) {
	sid, _ := c.Cookie("session_id")
	farmID := user_service.GetFarmFromToken(sid)

	name := strings.TrimSpace(c.PostForm("rascunhoNome"))
	if name == "" {
		toast := entity_public.GetWarningToast("Nome obrigatório", "Informe um nome para o rascunho")
		c.Header("HX-Trigger", toast.ToJsonStr())
		c.Status(http.StatusBadRequest)
		return
	}

	recipient, toast := parseDetachedProfileRecipient(c)
	if isFailureToast(toast) {
		c.Header("HX-Trigger", toast.ToJsonStr())
		c.Status(http.StatusBadRequest)
		return
	}

	items, itemsToast := parseDetachedProfileItems(c)
	if isFailureToast(itemsToast) {
		c.Header("HX-Trigger", itemsToast.ToJsonStr())
		c.Status(http.StatusBadRequest)
		return
	}

	rates, rateErr := parseUserTaxRates(c)
	if rateErr != nil {
		toast := entity_public.GetWarningToast("Falha ao identificar as alíquotas", rateErr.Error())
		c.Header("HX-Trigger", toast.ToJsonStr())
		c.Status(http.StatusBadRequest)
		return
	}

	// Updating an existing rascunho is an explicit action: the dialog submits
	// the applied rascunho id plus the "atualizarRascunho" confirmation.
	var profileID int
	if c.PostForm("atualizarRascunho") != "" {
		if id, err := strconv.Atoi(c.PostForm("rascunhoAtualId")); err == nil && id > 0 {
			profileID = id
		}
	}

	input := nfe_service.DetachedProfileInput{
		ID:         profileID,
		FarmID:     farmID,
		Name:       name,
		Recipient:  recipient,
		Items:      items,
		TaxRates:   rates,
		Overrides:  parseInvoiceOverrides(c),
		NaturezaOp: nil,
		ModFrete:   nil,
		InfCpl:     nil,
	}
	if overrides := input.Overrides; overrides != nil {
		input.NaturezaOp = overrides.NaturezaOp
		input.ModFrete = overrides.ModFrete
		input.InfCpl = overrides.InfCpl
	}

	svc := nfe_service.NewNFeService()
	_, saveToast := svc.SaveDetachedProfile(input)
	if isFailureToast(saveToast) {
		c.Header("HX-Trigger", saveToast.ToJsonStr())
		if saveToast.Type == entity_public.WarningToast {
			c.Status(http.StatusBadRequest)
		} else {
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	// 204: no swap, keep the emission form exactly as the user left it; the
	// success toast is driven by the HX-Trigger header.
	c.Header("HX-Trigger", saveToast.ToJsonStr())
	c.Status(http.StatusNoContent)
}

// UseDetachedProfileRoutes registers the Rascunho de NF-e management and
// save-from-form routes. The apply route (`/emitir/rascunho/:rascunhoId`) is
// registered with the detached emission routes.
func UseDetachedProfileRoutes(router gin.IRoutes) {
	router.GET("/rascunhos", getDetachedProfilesPage)
	router.GET("/rascunho/list", getDetachedProfileList)
	router.GET("/rascunho/form", getDetachedProfileForm)
	router.GET("/rascunho/form/:id", getFilledDetachedProfileForm)
	router.POST("/rascunho", addDetachedProfile)
	router.POST("/rascunho/salvar", saveDetachedProfileFromEmissionForm)
	router.PUT("/rascunho/:id", putDetachedProfile)
	router.DELETE("/rascunho/:id", deleteDetachedProfile)
}
