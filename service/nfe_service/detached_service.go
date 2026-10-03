package nfe_service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	entity_public "armazenda/entity/public"
	model_error "armazenda/model/error"
	"armazenda/model/farm_config_model"
	"armazenda/model/nfe_model"
	"armazenda/model/person_model"
	"armazenda/pkg/nfe/config"
	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"
	"armazenda/pkg/nfe/sefaz"
	"armazenda/pkg/nfe/service"
	nfe_xml "armazenda/pkg/nfe/xml"

	"github.com/shopspring/decimal"
)

// DetachedRecipient represents a recipient for detached NF-e
// Can be either an existing person or a new inline-defined recipient
type DetachedRecipient struct {
	// If PersonID is set, use existing person
	PersonID *uint32
	// Otherwise, create a new person with these fields
	// Type is inferred from Document (CPF → natural, CNPJ → legal)
	Name         string
	Document     string // CPF or CNPJ
	IE           *string
	Street       *string
	Number       *string
	Neighborhood *string
	City         *string
	State        *string
	CEP          *string
	PhoneNumber  *string
	Email        *string
}

// DetachedItemInput represents a single item in a detached invoice
type DetachedItemInput struct {
	FarmProductID *uint16 // nullable - if using a farm product
	// If FarmProductID is nil, use these fields directly
	ProductName string
	NCM         string
	CEST        *string
	Quantity    decimal.Decimal
	GrossWeight decimal.Decimal
	UnitPrice   decimal.Decimal
	Unit        string
}

// DetachedInvoiceInput holds all data needed to build a detached NF-e
type DetachedInvoiceInput struct {
	FarmID    uint32
	Recipient DetachedRecipient
	VehicleID *uint16 // optional
	CFOP      string
	TaxRates  entity.TaxRates
	Overrides *entity.InvoiceOverrides
	Items     []DetachedItemInput
}

// DetachedInvoiceResult is returned when a detached NF-e was built and
// submitted to SEFAZ. AccessKey identifies the persisted invoice so the
// caller can offer the signed XML for download.
type DetachedInvoiceResult struct {
	XML       string
	AccessKey string
}

// detachedBuildData carries the validated mapping produced by
// prepareDetachedBuildData. It is shared by the emission and preview paths;
// ItemsForDB, RecipientID and the tax totals are only consumed by the
// emission path.
type detachedBuildData struct {
	InvoiceInput entity.InvoiceInput
	ItemsForDB   []nfe_model.DetachedInvoiceItem
	FarmConfig   *entity_public.FarmConfig
	RecipientID  uint32
	TotalIBS     decimal.Decimal
	TotalCBS     decimal.Decimal
}

// prepareDetachedBuildData fetches the farm config, resolves the recipient,
// validates emitter/recipient/items and maps the request into the
// entity.InvoiceInput consumed by the XML builder. It is the single-source
// preparation step shared by the emission and preview paths.
//
// When dryRun is true it performs no writes: an inline recipient is mapped to
// RecipientData without creating a person row (RecipientID stays zero). The
// preview calls it with dryRun=true; the emission calls it with false.
func (s *NFeService) prepareDetachedBuildData(input DetachedInvoiceInput, dryRun bool) (detachedBuildData, entity_public.Toast) {
	nfeModel := nfe_model.GetNFeModel()
	empty := detachedBuildData{}

	// Get farm NFe config
	farmNFeConfig, dbErr := nfeModel.GetFarmConfig(input.FarmID)
	if dbErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("GetFarmConfig error: %v", dbErr.Error()))
		return empty, entity_public.GetErrorToast("Falha ao buscar configuração da NF-e", "")
	}
	if farmNFeConfig == nil {
		return empty, entity_public.GetWarningToast("NF-e não configurada", "acesse NF-e em Configurações")
	}

	// Get full farm data for the emitter address (same source as the
	// departure flow)
	farmData, farmErr := farm_config_model.GetFarmConfigModel().GetFarmConfig(input.FarmID)
	if farmErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("GetFarmConfig (farm data) error: %v", farmErr.Error()))
		return empty, entity_public.GetErrorToast("Falha ao buscar os dados da fazenda", "")
	}
	if farmData == nil {
		return empty, entity_public.GetWarningToast("Fazenda não encontrada", "complete o cadastro da fazenda")
	}

	// Resolve or create recipient
	recipientID, recipient, toast := s.resolveOrCreateRecipient(input.Recipient, input.FarmID, nfeModel, dryRun)
	if toast.Type == entity_public.ErrorToast || toast.Type == entity_public.WarningToast {
		return empty, toast
	}

	// Validate emitter
	emitter := s.mapFarmToEmitterForDetached(farmNFeConfig, farmData, nfeModel)
	validationErrors := s.validateEmitter(emitter, farmNFeConfig.Serie)
	if len(validationErrors) > 0 {
		return empty, entity_public.GetWarningToast(
			"Configuração de NF-e incompleta",
			strings.Join(validationErrors, "; "),
		)
	}

	// Validate recipient
	recipientErrors := validateRecipient(recipient)
	if len(recipientErrors) > 0 {
		return empty, entity_public.GetWarningToast(
			"Dados do destinatário incompletos",
			strings.Join(recipientErrors, "; "),
		)
	}

	// Resolve CFOP
	effectiveCFOP := input.CFOP
	if effectiveCFOP == "" {
		effectiveCFOP = farmNFeConfig.DefaultCFOP
	}

	// Build items
	items, itemsForDB, totalValue, totalIBS, totalCBS := s.buildDetachedItems(input.Items, effectiveCFOP, farmNFeConfig, input.TaxRates, input.Overrides)
	if len(items) == 0 {
		return empty, entity_public.GetWarningToast("Nenhum item informado", "Adicione pelo menos um item à NF-e")
	}

	// Resolve NaturezaOp
	naturezaOp := ""
	if input.Overrides != nil && input.Overrides.NaturezaOp != nil && *input.Overrides.NaturezaOp != "" {
		naturezaOp = *input.Overrides.NaturezaOp
	} else if farmNFeConfig.DefaultNaturezaOp != nil && *farmNFeConfig.DefaultNaturezaOp != "" {
		naturezaOp = *farmNFeConfig.DefaultNaturezaOp
	} else {
		naturezaOp = defaults.NaturezaOpForCFOP(effectiveCFOP)
	}

	// Resolve ModFrete
	effectiveModFrete := farmNFeConfig.DefaultModFrete
	if input.Overrides != nil && input.Overrides.ModFrete != nil {
		effectiveModFrete = *input.Overrides.ModFrete
	}

	// Build transport data
	transport := entity.TransportData{
		ModFrete: effectiveModFrete,
	}
	if input.VehicleID != nil && *input.VehicleID > 0 {
		// TODO: fetch vehicle plate
		// For now, skip vehicle data
	}

	// Calculate total weights from items
	var totalNetWeight, totalGrossWeight decimal.Decimal
	for _, item := range itemsForDB {
		totalNetWeight = totalNetWeight.Add(item.Quantity)
		totalGrossWeight = totalGrossWeight.Add(item.GrossWeight)
	}
	if !totalNetWeight.IsZero() || !totalGrossWeight.IsZero() {
		transport.Volumes = []entity.VolumeData{{
			QVol:  1,
			Esp:   "Granel",
			PesoL: totalNetWeight,
			PesoB: totalGrossWeight,
		}}
	}

	// Build infCpl (manual entry for detached)
	infCpl := ""
	if input.Overrides != nil && input.Overrides.InfCpl != nil {
		infCpl = *input.Overrides.InfCpl
	}

	// Build invoice input (number/CNF are allocated later by the emission
	// path; the preview consumes this as-is with Numero 0)
	invoiceInput := entity.InvoiceInput{
		Serie:       farmNFeConfig.Serie,
		Numero:      0,
		Environment: farmNFeConfig.Environment,
		TpEmis:      defaults.EmissaoNormal,
		NaturezaOp:  naturezaOp,
		Emitter:     emitter,
		Recipient:   recipient,
		Items:       items,
		Transport:   transport,
		Payment: entity.PaymentData{
			IndPag: 1,
			Detalhes: []entity.PagamentoDetalhe{
				{
					IndPag: 1,
					TPag:   "90",
					VPag:   totalValue,
				},
			},
		},
		TotalValue:            totalValue,
		InformacoesAdicionais: infCpl,
	}

	return detachedBuildData{
		InvoiceInput: invoiceInput,
		ItemsForDB:   itemsForDB,
		FarmConfig:   farmNFeConfig,
		RecipientID:  recipientID,
		TotalIBS:     totalIBS,
		TotalCBS:     totalCBS,
	}, entity_public.Toast{}
}

// BuildDetachedInvoice builds, signs, and attempts to send a detached NF-e
func (s *NFeService) BuildDetachedInvoice(input DetachedInvoiceInput) (DetachedInvoiceResult, entity_public.Toast) {
	nfeModel := nfe_model.GetNFeModel()

	build, toast := s.prepareDetachedBuildData(input, false)
	if toast.Type == entity_public.ErrorToast || toast.Type == entity_public.WarningToast {
		return DetachedInvoiceResult{}, toast
	}

	farmNFeConfig := build.FarmConfig
	invoiceInput := build.InvoiceInput

	// Values resolved during preparation, kept local for persistence below.
	effectiveCFOP := invoiceInput.Items[0].Produto.CFOP
	naturezaOp := invoiceInput.NaturezaOp
	effectiveModFrete := invoiceInput.Transport.ModFrete
	infCpl := invoiceInput.InformacoesAdicionais

	// Allocate number
	number, allocErr := nfeModel.AllocateDetachedNumber(input.FarmID, farmNFeConfig.Serie)
	if allocErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("AllocateDetachedNumber error: %v", allocErr.Error()))
		return DetachedInvoiceResult{}, entity_public.GetErrorToast("Falha ao alocar número da NF-e", "")
	}
	invoiceInput.Numero = number
	invoiceInput.CNF = generateRandomCNF()
	invoiceInput.TpEmis = defaults.EmissaoNormal

	// Generate access key
	accessKey := entity.GenerateAccessKey(entity.AccessKeyData{
		CUF:      defaults.UFCode(invoiceInput.Emitter.UF),
		AAMM:     time.Now().Format("0601"),
		Document: s.documentForAccessKey(invoiceInput.Emitter),
		Mod:      defaults.ModeloNFe,
		Serie:    invoiceInput.Serie,
		NNF:      invoiceInput.Numero,
		TpEmis:   defaults.EmissaoNormal.String(),
		CNF:      invoiceInput.CNF,
	})

	// Sign XML
	certPassword := decryptPassword(farmNFeConfig.CertificatePasswordEncrypted)
	sefazCfg := config.SefazConfig{
		Environment: config.Environment(farmNFeConfig.Environment),
		StateUF:     farmNFeConfig.EmitterUF,
		Timeout:     30 * time.Second,
	}
	invService := service.NewInvoiceService(sefazCfg)

	signedXML, signErr := invService.BuildAndSign(invoiceInput, farmNFeConfig.CertificateData, certPassword)
	if signErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("BuildAndSign error: %v", signErr.Error()))
		return DetachedInvoiceResult{}, entity_public.GetErrorToast("Falha ao construir e assinar NF-e", signErr.Error())
	}

	// Persist invoice
	var ratesToPersist *entity.TaxRates
	if hasAnyUserRate(input.TaxRates) {
		ratesToPersist = &input.TaxRates
	}
	if input.Overrides == nil {
		input.Overrides = &entity.InvoiceOverrides{}
	}
	if input.Overrides.InfCpl == nil {
		input.Overrides.InfCpl = &infCpl
	}

	invoiceID, createErr := nfeModel.CreateDetachedInvoice(
		input.FarmID, build.RecipientID, accessKey, farmNFeConfig.Serie, number,
		effectiveCFOP, &naturezaOp, &effectiveModFrete,
		invoiceInput.TotalValue, build.TotalIBS, build.TotalCBS,
		1, ratesToPersist, input.Overrides, build.ItemsForDB,
	)
	if createErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("CreateDetachedInvoice error: %v", createErr.Error()))
		return DetachedInvoiceResult{}, entity_public.GetErrorToast("Falha ao salvar NF-e", "")
	}

	// Store signed XML
	if xmlErr := nfeModel.UpdateDetachedInvoiceSignedXML(invoiceID, signedXML); xmlErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateDetachedInvoiceSignedXML error: %v", xmlErr.Error()))
	}

	// Send to SEFAZ
	sefazResp, sendErr := invService.SendToSefaz(signedXML, farmNFeConfig.CertificateData, certPassword)
	if sendErr == nil && sefazResp != nil {
		return s.handleDetachedSefazResponse(sefazResp, invoiceID, signedXML, accessKey, nfeModel)
	}

	// SEFAZ failed - check SVC
	svcResp, svcErr := invService.CheckSVCStatus(farmNFeConfig.CertificateData, certPassword)
	if svcErr == nil && svcResp != nil && svcResp.IsSVCOperational() {
		return s.attemptDetachedSVCContingency(invoiceInput, input, farmNFeConfig, signedXML, invoiceID, accessKey, nfeModel, invService, certPassword, ratesToPersist, build.ItemsForDB)
	}

	// Both SEFAZ and SVC unavailable
	if sendErr != nil {
		errUpd := nfeModel.UpdateDetachedInvoiceStatus(invoiceID, "draft", "", "", "SEFAZ e SVC indisponiveis: "+sendErr.Error())
		if errUpd != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateDetachedInvoiceStatus error: %v", errUpd.Error()))
		}
	} else {
		errUpd := nfeModel.UpdateDetachedInvoiceStatus(invoiceID, "draft", "", "", "SEFAZ e SVC indisponiveis: resposta vazia")
		if errUpd != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateDetachedInvoiceStatus error: %v", errUpd.Error()))
		}
	}
	return DetachedInvoiceResult{}, entity_public.GetErrorToast(
		"SEFAZ e SVC indisponiveis",
		"NF-e nao pode ser emitida no momento. Tente novamente mais tarde.",
	)
}

// resolveOrCreateRecipient resolves an existing recipient or maps a new
// inline recipient. When dryRun is false the inline recipient is also
// persisted as a person row; when true (preview) only the mapping runs and
// the returned recipientID is zero.
func (s *NFeService) resolveOrCreateRecipient(input DetachedRecipient, farmID uint32, nfeModel *nfe_model.NFeModel, dryRun bool) (uint32, entity.RecipientData, entity_public.Toast) {
	// If PersonID is set, fetch existing person
	if input.PersonID != nil {
		pModel := person_model.GetPersonModel()
		person, personErr := pModel.GetFullPersonById(*input.PersonID)
		if personErr != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("GetFullPersonById error: %v", personErr.Error()))
			return 0, entity.RecipientData{}, entity_public.GetErrorToast("Falha ao buscar destinatário", "")
		}
		recipient := s.mapFullPersonToRecipient(person, nfeModel)
		return *input.PersonID, recipient, entity_public.Toast{}
	}

	// Map the inline data to the XML recipient struct (no writes).
	recipient := s.mapInlineRecipientData(input, nfeModel)

	// Preview: dry-run only, never create the person row.
	if dryRun {
		return 0, recipient, entity_public.Toast{}
	}

	// Create person record (emission only)
	pModel := person_model.GetPersonModel()
	recipientInput := person_model.DetachedRecipientInput{
		Name:         input.Name,
		Document:     input.Document,
		IE:           input.IE,
		Street:       input.Street,
		Number:       input.Number,
		Neighborhood: input.Neighborhood,
		City:         input.City,
		State:        input.State,
		CEP:          input.CEP,
		PhoneNumber:  input.PhoneNumber,
		Email:        input.Email,
	}
	personID, createErr := pModel.CreatePersonForDetachedNFe(farmID, detachedDocumentType(input.Document), recipientInput)
	if createErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("CreatePersonForDetachedNFe error: %v", createErr.Error()))
		return 0, entity.RecipientData{}, entity_public.GetErrorToast("Falha ao criar destinatário", "")
	}

	return personID, recipient, entity_public.Toast{}
}

// mapInlineRecipientData maps the inline recipient form data to the
// entity.RecipientData consumed by the XML builder. It performs no writes;
// the only database access is the municipality-code lookup.
func (s *NFeService) mapInlineRecipientData(input DetachedRecipient, nfeModel *nfe_model.NFeModel) entity.RecipientData {
	recipient := entity.RecipientData{
		Type:       1, // default CNPJ
		XNome:      input.Name,
		Logradouro: "",
		Numero:     "",
		Bairro:     "",
		CodigoMun:  "",
		Municipio:  "",
		UF:         "",
		CEP:        "",
		Fone:       "",
		IndIEDest:  "9",
	}

	if detachedDocumentType(input.Document) == 2 {
		recipient.Type = 1 // CNPJ
		recipient.CNPJ = input.Document
	} else {
		recipient.Type = 2 // CPF
		recipient.CPF = input.Document
	}

	if input.IE != nil && *input.IE != "" {
		recipient.IE = *input.IE
		recipient.IndIEDest = "1"
	}
	if input.Street != nil {
		recipient.Logradouro = *input.Street
	}
	if input.Number != nil {
		recipient.Numero = *input.Number
	}
	if recipient.Numero == "" {
		recipient.Numero = "S/N"
	}
	if input.Neighborhood != nil {
		recipient.Bairro = *input.Neighborhood
	}
	if input.City != nil {
		recipient.Municipio = *input.City
	}
	if input.State != nil {
		recipient.UF = *input.State
	}
	if input.CEP != nil {
		recipient.CEP = *input.CEP
	}
	if input.PhoneNumber != nil {
		recipient.Fone = *input.PhoneNumber
	}

	// Resolve municipality code
	if input.City != nil && input.State != nil {
		code, _ := nfeModel.GetMunicipio(*input.City, *input.State)
		if code != "" {
			recipient.CodigoMun = code
		}
	}

	return recipient
}

// detachedDocumentType infers the person type from the document length:
// 2 (legal / CNPJ) above 11 characters, 1 (natural / CPF) otherwise.
func detachedDocumentType(document string) int {
	if len(document) > 11 {
		return 2
	}
	return 1
}

// mapFarmToEmitterForDetached maps farm config and farm address data to
// emitter data for detached invoices. It delegates to the departure-flow
// mapper so both emissions produce identical emitter groups.
func (s *NFeService) mapFarmToEmitterForDetached(cfg *entity_public.FarmConfig, farm *entity_public.Farm, nfeModel *nfe_model.NFeModel) entity.EmitterData {
	return s.mapFarmToEmitter(cfg, farm, nfeModel)
}

// buildDetachedItems builds item data for detached invoices
func (s *NFeService) buildDetachedItems(inputs []DetachedItemInput, defaultCFOP string, farmConfig *entity_public.FarmConfig, userRates entity.TaxRates, overrides *entity.InvoiceOverrides) ([]entity.ItemData, []nfe_model.DetachedInvoiceItem, decimal.Decimal, decimal.Decimal, decimal.Decimal) {
	var items []entity.ItemData
	var itemsForDB []nfe_model.DetachedInvoiceItem
	var totalValue, totalIBS, totalCBS decimal.Decimal

	rates := MergeRates(userRates, farmConfig)

	for _, input := range inputs {
		// Determine product details
		productName := input.ProductName
		ncm := input.NCM
		cest := input.CEST
		unit := input.Unit

		if unit == "" {
			unit = farmConfig.DefaultUnit
			if unit == "" {
				unit = "KG"
			}
		}

		if overrides != nil {
			if overrides.ProductDesc != nil && *overrides.ProductDesc != "" {
				productName = *overrides.ProductDesc
			}
			if overrides.NCM != nil && *overrides.NCM != "" {
				ncm = *overrides.NCM
			}
			if overrides.CEST != nil {
				cest = overrides.CEST
			}
			if overrides.Unit != nil && *overrides.Unit != "" {
				unit = *overrides.Unit
			}
		}

		if ncm == "" {
			ncm = defaults.NCMSoja
		}
		if productName == "" {
			productName = "Produto Agricola"
		}

		// Calculate total
		itemTotal := input.Quantity.Mul(input.UnitPrice)
		totalValue = totalValue.Add(itemTotal)

		// Determine tax CSTs
		regime := defaults.TaxRegime(farmConfig.TaxRegime)
		var icmsCST, icmsCSOSN string
		var vBC, vICMS decimal.Decimal
		var picms decimal.Decimal

		if regime == defaults.TaxRegimeSimplesNacional {
			if overrides != nil && overrides.ICMSCST != nil && *overrides.ICMSCST != "" {
				icmsCSOSN = *overrides.ICMSCST
			} else if farmConfig.DefaultICMSCST != nil && *farmConfig.DefaultICMSCST != "" {
				icmsCSOSN = *farmConfig.DefaultICMSCST
			} else {
				icmsCSOSN = defaults.CSOSNSemPermissaoCredito
			}
			vBC = decimal.Zero
			vICMS = decimal.Zero
			picms = decimal.Zero
		} else {
			if overrides != nil && overrides.ICMSCST != nil && *overrides.ICMSCST != "" {
				icmsCST = *overrides.ICMSCST
			} else if farmConfig.DefaultICMSCST != nil && *farmConfig.DefaultICMSCST != "" {
				icmsCST = *farmConfig.DefaultICMSCST
			} else {
				icmsCST = defaults.CSTTributadaIntegral
			}
			vBC = itemTotal
			vICMS = itemTotal.Mul(*rates.ICMSRate)
			picms = rates.ICMSRate.Mul(decimal.NewFromInt(100))
		}

		// PIS CST
		pisCST := "01"
		if overrides != nil && overrides.PISCST != nil && *overrides.PISCST != "" {
			pisCST = *overrides.PISCST
		} else if farmConfig.DefaultPISCST != nil && *farmConfig.DefaultPISCST != "" {
			pisCST = *farmConfig.DefaultPISCST
		}

		// COFINS CST
		cofinsCST := "01"
		if overrides != nil && overrides.COFINSCST != nil && *overrides.COFINSCST != "" {
			cofinsCST = *overrides.COFINSCST
		} else if farmConfig.DefaultCOFINSCST != nil && *farmConfig.DefaultCOFINSCST != "" {
			cofinsCST = *farmConfig.DefaultCOFINSCST
		}

		// IBS/CBS
		ibsCbsCST := defaults.IBSCBSCSTTributadaIntegral
		if overrides != nil && overrides.CBSCST != nil && *overrides.CBSCST != "" {
			ibsCbsCST = *overrides.CBSCST
		} else if overrides != nil && overrides.IBSCST != nil && *overrides.IBSCST != "" {
			ibsCbsCST = *overrides.IBSCST
		}
		cClassTrib := defaults.CClassTribDefault
		if overrides != nil && overrides.CClassTrib != nil && *overrides.CClassTrib != "" {
			cClassTrib = *overrides.CClassTrib
		}

		var ibsVBC, ibsVIBSUF, ibsVIBSMun, ibsVIBS, ibsVCBS decimal.Decimal
		if defaults.IsTaxReformActive(time.Now()) {
			ibsVBC = itemTotal
			ibsVIBSUF = itemTotal.Mul(*rates.IBSRate)
			ibsVIBSMun = decimal.Zero
			ibsVIBS = ibsVIBSUF.Add(ibsVIBSMun)
			ibsVCBS = itemTotal.Mul(*rates.CBSRate)
			totalIBS = totalIBS.Add(ibsVIBS)
			totalCBS = totalCBS.Add(ibsVCBS)
		}

		itemData := entity.ItemData{
			Numero: len(items) + 1,
			Produto: entity.ProdutoData{
				Codigo:   fmt.Sprintf("%d", len(items)+1),
				CEAN:     "SEM GTIN",
				XProd:    productName,
				NCM:      ncm,
				CEST:     safeStringPtr(cest),
				CFOP:     defaultCFOP,
				UCom:     unit,
				QCom:     input.Quantity,
				VUnCom:   input.UnitPrice,
				VProd:    itemTotal,
				CEANTrib: "SEM GTIN",
				UTrib:    unit,
				QTrib:    input.Quantity,
				VUnTrib:  input.UnitPrice,
				IndTot:   1,
			},
			Imposto: entity.ImpostoData{
				ICMS: entity.ICMSData{
					Origem: defaults.ICMSOrigemNacional,
					CST:    icmsCST,
					CSOSN:  icmsCSOSN,
					ModBC:  "3",
					VBC:    vBC,
					PICMS:  picms,
					VICMS:  vICMS,
				},
				PIS: entity.PISData{
					CST:  pisCST,
					VBC:  itemTotal,
					PPIS: rates.PISRate.Mul(decimal.NewFromInt(100)),
					VPIS: itemTotal.Mul(*rates.PISRate),
				},
				COFINS: entity.COFINSData{
					CST:     cofinsCST,
					VBC:     itemTotal,
					PCOFINS: rates.COFINSRate.Mul(decimal.NewFromInt(100)),
					VCOFINS: itemTotal.Mul(*rates.COFINSRate),
				},
				IBSCBS: entity.IBSCBSData{
					CST:        ibsCbsCST,
					CClassTrib: cClassTrib,
					VBC:        ibsVBC,
					VIBSUF:     ibsVIBSUF,
					VIBSMun:    ibsVIBSMun,
					VIBS:       ibsVIBS,
					PCBS:       rates.CBSRate.Mul(decimal.NewFromInt(100)),
					VCBS:       ibsVCBS,
				},
			},
		}

		items = append(items, itemData)

		// Build item for DB storage
		itemForDB := nfe_model.DetachedInvoiceItem{
			FarmProductID: input.FarmProductID,
			ProductName:   productName,
			NCM:           ncm,
			CEST:          cest,
			CFOP:          defaultCFOP,
			Unit:          unit,
			Quantity:      input.Quantity,
			GrossWeight:   input.GrossWeight,
			UnitPrice:     input.UnitPrice,
			TotalValue:    itemTotal,
		}
		itemsForDB = append(itemsForDB, itemForDB)
	}

	return items, itemsForDB, totalValue, totalIBS, totalCBS
}

// GenerateDetachedPreviewDANFE builds a non-fiscal preview DANFE PDF for a
// detached NF-e. It runs the same validation as the emission path but performs
// no side effects: no number allocation, no person creation, no signing and no
// persistence.
func (s *NFeService) GenerateDetachedPreviewDANFE(input DetachedInvoiceInput) ([]byte, entity_public.Toast) {
	build, toast := s.prepareDetachedBuildData(input, true)
	if toast.Type == entity_public.ErrorToast || toast.Type == entity_public.WarningToast {
		return nil, toast
	}

	data := s.buildDetachedDANFEData(build.InvoiceInput)

	generator := service.NewDANFEGenerator()
	pdfBytes, genErr := generator.GeneratePreview(data)
	if genErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("GeneratePreview (detached) error: %v", genErr.Error()))
		return nil, entity_public.GetErrorToast("Falha ao gerar a pré-visualização da DANFE", genErr.Error())
	}

	return pdfBytes, entity_public.Toast{}
}

// buildDetachedDANFEData maps a prepared InvoiceInput to the DANFEData used
// for the non-fiscal preview PDF. All items are mapped (not just the first,
// unlike the departure flow) and invoice-level totals are summed across them
// so the preview shows the same values the emitted XML will carry.
func (s *NFeService) buildDetachedDANFEData(input entity.InvoiceInput) entity.DANFEData {
	now := time.Now()

	tpAmb := "1"
	if input.Environment == 2 {
		tpAmb = "2"
	}

	products := make([]entity.DANFEProduct, 0, len(input.Items))
	var vbc, vicms, vpis, vcofins, vbcIBSCBS, vibs, vcbs decimal.Decimal
	for _, item := range input.Items {
		product := item.Produto
		imp := item.Imposto

		// Simples Nacional emits CSOSN instead of CST; the DANFE column is
		// shared, so show whichever code the emitted XML will carry.
		icmsCode := imp.ICMS.CST
		if icmsCode == "" {
			icmsCode = imp.ICMS.CSOSN
		}

		products = append(products, entity.DANFEProduct{
			Code:      product.Codigo,
			Desc:      product.XProd,
			NCM:       product.NCM,
			CST:       icmsCode,
			CFOP:      product.CFOP,
			Unit:      product.UCom,
			Quantity:  product.QCom,
			UnitPrice: product.VUnCom,
			Total:     product.VProd,
			UTrib:     product.UTrib,
			QTrib:     product.QTrib,
			VUnTrib:   product.VUnTrib,
			InfAdProd: item.InfAdProd,
			VBC:       imp.ICMS.VBC,
			PICMS:     imp.ICMS.PICMS,
			VICMS:     imp.ICMS.VICMS,
			PPIS:      imp.PIS.PPIS,
			VPIS:      imp.PIS.VPIS,
			PCOFINS:   imp.COFINS.PCOFINS,
			VCOFINS:   imp.COFINS.VCOFINS,
			// PIBS is derived from VIBSUF / VBC * 100 since the per-item
			// rate is no longer persisted on the entity (only the values).
			PIBS: perItemRate(imp.IBSCBS.VIBSUF, imp.IBSCBS.VBC),
			VIBS: imp.IBSCBS.VIBS,
			PCBS: imp.IBSCBS.PCBS,
			VCBS: imp.IBSCBS.VCBS,
		})

		vbc = vbc.Add(imp.ICMS.VBC)
		vicms = vicms.Add(imp.ICMS.VICMS)
		vpis = vpis.Add(imp.PIS.VPIS)
		vcofins = vcofins.Add(imp.COFINS.VCOFINS)
		vbcIBSCBS = vbcIBSCBS.Add(imp.IBSCBS.VBC)
		vibs = vibs.Add(imp.IBSCBS.VIBS)
		vcbs = vcbs.Add(imp.IBSCBS.VCBS)
	}

	data := entity.DANFEData{
		AccessKey:           "",
		Numero:              0,
		Serie:               input.Serie,
		NaturezaOp:          input.NaturezaOp,
		EmissionDate:        now.Format("02/01/2006 15:04:05"),
		TpEmis:              input.TpEmis.String(),
		TpAmb:               tpAmb,
		TpNF:                "1", // saída (matches builder.go:95)
		EmitterName:         input.Emitter.XNome,
		EmitterCNPJ:         s.formatDocument(input.Emitter.Document, input.Emitter.Document),
		EmitterIE:           input.Emitter.IE,
		EmitterCRT:          input.Emitter.CRT,
		EmitterAddress:      input.Emitter.Logradouro,
		EmitterNumber:       input.Emitter.Numero,
		EmitterNeighborhood: input.Emitter.Bairro,
		EmitterCEP:          input.Emitter.CEP,
		EmitterCity:         input.Emitter.Municipio,
		EmitterUF:           input.Emitter.UF,
		EmitterPhone:        input.Emitter.Fone,
		DestName:            input.Recipient.XNome,
		DestCNPJ:            s.formatDocument(input.Recipient.CNPJ, input.Recipient.CPF),
		DestIE:              input.Recipient.IE,
		DestIndIEDest:       input.Recipient.IndIEDest,
		DestAddress:         input.Recipient.Logradouro,
		DestNumber:          input.Recipient.Numero,
		DestNeighborhood:    input.Recipient.Bairro,
		DestCEP:             input.Recipient.CEP,
		DestCity:            input.Recipient.Municipio,
		DestUF:              input.Recipient.UF,
		DestPhone:           input.Recipient.Fone,
		Products:            products,
		TotalValue:          input.TotalValue,
		VBC:                 vbc,
		VICMS:               vicms,
		VPIS:                vpis,
		VCOFINS:             vcofins,
		VBCIBSCBS:           vbcIBSCBS,
		VIBS:                vibs,
		VCBS:                vcbs,
		ModFrete:            strconv.Itoa(input.Transport.ModFrete),
		InfCpl:              input.InformacoesAdicionais,
	}

	transport := input.Transport
	if transport.Transportadora != nil {
		data.TranspName = transport.Transportadora.XNome
		data.TranspCNPJ = s.formatDocument(transport.Transportadora.CNPJ, transport.Transportadora.CPF)
		data.TranspIE = transport.Transportadora.IE
		data.TranspAddress = transport.Transportadora.Endereco
		data.TranspCity = transport.Transportadora.Municipio
		data.TranspUF = transport.Transportadora.UF
	}
	if len(transport.Volumes) > 0 {
		vol := transport.Volumes[0]
		data.QVol = strconv.Itoa(vol.QVol)
		data.Esp = vol.Esp
		data.Marca = vol.Marca
		data.NVol = vol.NVol
		data.PesoL = vol.PesoL
		data.PesoB = vol.PesoB
	}
	if transport.Veiculo != nil {
		data.VeicPlate = transport.Veiculo.Placa
		data.VeicUF = transport.Veiculo.UF
	}

	return data
}

// safeStringPtr returns the string pointed to by p, or empty string if nil
func safeStringPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// handleDetachedSefazResponse processes the SEFAZ response for detached invoices
func (s *NFeService) handleDetachedSefazResponse(sefazResp *sefaz.AutorizacaoResponse, invoiceID int, signedXML, accessKey string, nfeModel *nfe_model.NFeModel) (DetachedInvoiceResult, entity_public.Toast) {
	result := DetachedInvoiceResult{XML: signedXML, AccessKey: accessKey}

	if sefazResp.IsAuthorized() {
		errUpd := nfeModel.UpdateDetachedInvoiceStatus(invoiceID, "authorized", sefazResp.Protocol, sefazResp.StatusCode, sefazResp.StatusMotive)
		if errUpd != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateDetachedInvoiceStatus error: %v", errUpd.Error()))
		}
		return result, entity_public.GetSuccessToast("NF-e autorizada pela SEFAZ", fmt.Sprintf("Protocolo: %s", sefazResp.Protocol))
	}

	if sefazResp.IsProcessing() || sefazResp.IsAccepted() {
		errUpd := nfeModel.UpdateDetachedInvoiceStatus(invoiceID, "pending", sefazResp.Protocol, sefazResp.StatusCode, sefazResp.StatusMotive)
		if errUpd != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateDetachedInvoiceStatus error: %v", errUpd.Error()))
		}
		return result, entity_public.GetSuccessToast("NF-e enviada à SEFAZ e em processamento", fmt.Sprintf("Status: %s - %s", sefazResp.StatusCode, sefazResp.StatusMotive))
	}

	if sefazResp.IsRejected() {
		errUpd := nfeModel.UpdateDetachedInvoiceStatus(invoiceID, "denied", "", sefazResp.StatusCode, sefazResp.StatusMotive)
		if errUpd != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateDetachedInvoiceStatus error: %v", errUpd.Error()))
		}
		return result, entity_public.GetErrorToast(
			fmt.Sprintf("NF-e rejeitada pela SEFAZ (%s)", sefazResp.StatusCode),
			sefazResp.StatusMotive,
		)
	}

	errUpd := nfeModel.UpdateDetachedInvoiceStatus(invoiceID, "pending", "", sefazResp.StatusCode, sefazResp.StatusMotive)
	if errUpd != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateDetachedInvoiceStatus error: %v", errUpd.Error()))
	}
	return result, entity_public.GetSuccessToast("NF-e enviada à SEFAZ e em processamento", fmt.Sprintf("Status: %s", sefazResp.StatusMotive))
}

// attemptDetachedSVCContingency tries to send the detached NF-e via SVC
func (s *NFeService) attemptDetachedSVCContingency(input entity.InvoiceInput, detachedInput DetachedInvoiceInput, farmNFeConfig *entity_public.FarmConfig, oldSignedXML string, oldInvoiceID int, oldAccessKey string, nfeModel *nfe_model.NFeModel, invService *service.InvoiceService, certPassword string, taxRates *entity.TaxRates, itemsForDB []nfe_model.DetachedInvoiceItem) (DetachedInvoiceResult, entity_public.Toast) {
	// Allocate new number
	newNumber, allocErr := nfeModel.AllocateDetachedNumber(detachedInput.FarmID, farmNFeConfig.Serie)
	if allocErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("AllocateDetachedNumber (SVC) error: %v", allocErr.Error()))
		return DetachedInvoiceResult{}, entity_public.GetErrorToast("Falha ao alocar número para contingência", "")
	}

	tpEmis := defaults.SVCForState(farmNFeConfig.EmitterUF)
	reason := "Indisponibilidade do ambiente de autorizacao da SEFAZ de origem"

	newSignedXML, newAccessKey, rebuildErr := invService.RebuildForContingency(input, newNumber, generateRandomCNF(), tpEmis, reason, farmNFeConfig.CertificateData, certPassword)
	if rebuildErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("RebuildForContingency error: %v", rebuildErr.Error()))
		return DetachedInvoiceResult{}, entity_public.GetErrorToast("Falha ao reconstruir NF-e para SVC", rebuildErr.Error())
	}

	// Calculate totals
	var totalValue, totalIBS, totalCBS decimal.Decimal
	for _, item := range itemsForDB {
		totalValue = totalValue.Add(item.TotalValue)
	}
	for _, item := range input.Items {
		totalIBS = totalIBS.Add(item.Imposto.IBSCBS.VIBS)
		totalCBS = totalCBS.Add(item.Imposto.IBSCBS.VCBS)
	}

	// Save new invoice
	// Get recipient ID from the original invoice
	originalInvoice, _ := nfeModel.GetDetachedInvoiceByAccessKey(oldAccessKey)
	recipientID := uint32(0)
	if originalInvoice != nil {
		recipientID = originalInvoice.RecipientID
	}
	newInvoiceID, createErr := nfeModel.CreateDetachedInvoice(
		detachedInput.FarmID, recipientID,
		newAccessKey, farmNFeConfig.Serie, newNumber,
		input.Items[0].Produto.CFOP, &input.NaturezaOp, nil,
		totalValue, totalIBS, totalCBS,
		int(tpEmis), taxRates, detachedInput.Overrides, itemsForDB,
	)
	if createErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("CreateDetachedInvoice (SVC) error: %v", createErr.Error()))
		return DetachedInvoiceResult{}, entity_public.GetErrorToast("Falha ao salvar NF-e de contingência", "")
	}

	xmlErr := nfeModel.UpdateDetachedInvoiceSignedXML(newInvoiceID, newSignedXML)
	if xmlErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateDetachedInvoiceSignedXML (SVC) error: %v", xmlErr.Error()))
		return DetachedInvoiceResult{}, entity_public.GetErrorToast("Falha ao salvar XML assinado de contingência", "")
	}

	// Send to SVC
	sefazResp, sendErr := invService.SendToSefazWithEmission(newSignedXML, farmNFeConfig.CertificateData, certPassword, tpEmis)
	if sendErr != nil {
		errUpd := nfeModel.UpdateDetachedInvoiceStatus(newInvoiceID, "draft", "", "", "SVC indisponivel: "+sendErr.Error())
		if errUpd != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateDetachedInvoiceStatus (SVC) error: %v", errUpd.Error()))
		}
		_ = nfeModel.SupersedeDetachedInvoice(oldInvoiceID, newInvoiceID)
		return DetachedInvoiceResult{}, entity_public.GetErrorToast(
			"SEFAZ e SVC indisponiveis",
			"NF-e nao pode ser emitida no momento. Tente novamente mais tarde.",
		)
	}

	if sefazResp == nil {
		errUpd := nfeModel.UpdateDetachedInvoiceStatus(newInvoiceID, "draft", "", "", "SVC indisponivel: resposta vazia")
		if errUpd != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateDetachedInvoiceStatus (SVC) error: %v", errUpd.Error()))
		}
		_ = nfeModel.SupersedeDetachedInvoice(oldInvoiceID, newInvoiceID)
		return DetachedInvoiceResult{}, entity_public.GetErrorToast(
			"SEFAZ e SVC indisponiveis",
			"NF-e nao pode ser emitida no momento. Tente novamente mais tarde.",
		)
	}

	// SVC responded
	result, resultToast := s.handleDetachedSefazResponse(sefazResp, newInvoiceID, newSignedXML, newAccessKey, nfeModel)
	_ = nfeModel.SupersedeDetachedInvoice(oldInvoiceID, newInvoiceID)
	return result, resultToast
}

// CancelDetachedInvoice cancels a detached NF-e
func (s *NFeService) CancelDetachedInvoice(accessKey, justification string, farmID uint32) entity_public.Toast {
	nfeModel := nfe_model.GetNFeModel()
	invoice, err := nfeModel.GetDetachedInvoiceByAccessKey(accessKey)
	if err != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("CancelDetachedInvoice GetDetachedInvoiceByAccessKey error: %v", err.Error()))
		return entity_public.GetErrorToast("Erro interno ao buscar a NF-e", "")
	}
	if invoice == nil {
		return entity_public.GetWarningToast("NF-e não encontrada", "")
	}

	if invoice.FarmID != farmID {
		return entity_public.GetWarningToast("Esta NF-e não pertence à sua fazenda", "")
	}

	if invoice.Status != "authorized" {
		return entity_public.GetWarningToast(
			"Esta NF-e não pode ser cancelada",
			"Somente NF-e autorizadas pela SEFAZ podem ser canceladas")
	}
	if invoice.Protocol == nil || *invoice.Protocol == "" {
		return entity_public.GetWarningToast(
			"Protocolo de autorização não encontrado",
			"Consulte a NF-e na SEFAZ e tente novamente")
	}

	justification = strings.TrimSpace(justification)
	if len([]rune(justification)) < 15 || len([]rune(justification)) > 256 {
		return entity_public.GetWarningToast(
			"Justificativa inválida",
			"A justificativa deve ter entre 15 e 256 caracteres")
	}

	farmNFeConfig, dbErr := nfeModel.GetFarmConfig(farmID)
	if dbErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("CancelDetachedInvoice GetFarmConfig error: %v", dbErr.Error()))
		return entity_public.GetErrorToast("Erro interno ao buscar a configuração da NF-e", "")
	}
	if farmNFeConfig == nil {
		return entity_public.GetWarningToast("NF-e não configurada", "acesse NF-e em Configurações")
	}

	certPassword := decryptPassword(farmNFeConfig.CertificatePasswordEncrypted)
	sefazCfg := config.SefazConfig{
		Environment: config.Environment(farmNFeConfig.Environment),
		StateUF:     farmNFeConfig.EmitterUF,
		Timeout:     30 * time.Second,
	}
	invService := service.NewInvoiceService(sefazCfg)

	emitterDoc := ""
	if farmNFeConfig.DocEmitter != nil {
		emitterDoc = *farmNFeConfig.DocEmitter
	}
	eventInput := nfe_xml.CancelEventInput{
		AccessKey:     invoice.AccessKey,
		Protocol:      *invoice.Protocol,
		Justification: justification,
		EmitterDoc:    emitterDoc,
		EmitterType:   farmNFeConfig.EmitterType,
		EmitterUF:     farmNFeConfig.EmitterUF,
		Environment:   farmNFeConfig.Environment,
		DhEvento:      time.Now(),
		SeqEvento:     1,
	}

	signedEventXML, buildErr := invService.BuildAndSignCancellationEvent(eventInput, farmNFeConfig.CertificateData, certPassword)
	if buildErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("BuildAndSignCancellationEvent error: %v", buildErr.Error()))
		return entity_public.GetErrorToast("Falha ao montar o evento de cancelamento", buildErr.Error())
	}

	tpEmis := defaults.TpEmis(invoice.TpEmis)
	resp, sendErr := invService.SendCancellationEvent(signedEventXML, farmNFeConfig.CertificateData, certPassword, tpEmis)
	if sendErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("SendCancellationEvent error: %v", sendErr.Error()))
		return entity_public.GetErrorToast(
			"Não foi possível comunicar com a SEFAZ",
			"O cancelamento não foi registrado. Tente novamente mais tarde.")
	}
	if resp == nil {
		return entity_public.GetErrorToast(
			"Não foi possível comunicar com a SEFAZ",
			"Resposta vazia. O cancelamento não foi registrado.")
	}

	if resp.IsRegistered() || resp.IsAlreadyCancelled() {
		updErr := nfeModel.UpdateDetachedInvoiceCancelled(invoice.ID, justification, signedEventXML, resp.StatusCode, resp.StatusMotive)
		if updErr != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateDetachedInvoiceCancelled error: %v", updErr.Error()))
			return entity_public.GetErrorToast(
				"Cancelamento registrado na SEFAZ, mas falhamos ao atualizar a NF-e localmente",
				"")
		}
		if resp.IsAlreadyCancelled() {
			return entity_public.GetSuccessToast(
				"NF-e já estava cancelada na SEFAZ",
				"A situação da NF-e foi atualizada")
		}
		return entity_public.GetSuccessToast(
			"NF-e cancelada",
			fmt.Sprintf("Motivo registrado: %s", justification))
	}

	model_error.GetLoggerModel().Log(fmt.Sprintf("SEFAZ cancelamento: status %s | motivo: %s", resp.StatusCode, resp.StatusMotive))
	return entity_public.GetWarningToast(
		fmt.Sprintf("SEFAZ rejeitou o cancelamento (%s)", resp.StatusCode),
		resp.StatusMotive,
	)
}
