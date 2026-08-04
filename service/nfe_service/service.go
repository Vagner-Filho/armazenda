package nfe_service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	entity_public "armazenda/entity/public"
	"armazenda/model/departure_model"
	model_error "armazenda/model/error"
	"armazenda/model/farm_config_model"
	"armazenda/model/nfe_model"
	"armazenda/model/person_model"
	"armazenda/model/product_model"
	"armazenda/model/vehicle_model"
	"armazenda/pkg/nfe/config"
	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"
	"armazenda/pkg/nfe/sefaz"
	"armazenda/pkg/nfe/service"
	nfe_xml "armazenda/pkg/nfe/xml"

	"github.com/shopspring/decimal"
)

// NFeService bridges Armazenda entities with the pkg/nfe package.
type NFeService struct{}

// NewNFeService creates a new NFe service.
func NewNFeService() *NFeService {
	return &NFeService{}
}

// prepareInvoiceBuildData fetches and validates all data needed to build an NF-e
// without allocating a number, signing, or persisting anything.
func (s *NFeService) prepareInvoiceBuildData(departureID uint32, unitPrice decimal.Decimal, farmID uint32, cfop string, userRates entity.TaxRates) (entity.InvoiceInput, entity_public.Departure, *entity_public.FarmConfig, entity_public.Product, entity_public.Toast) {
	var emptyInput entity.InvoiceInput

	// Get departure
	dModel := departure_model.GetDepartureModel()
	departure, err := dModel.GetDeparture(departureID)
	if err != nil {
		if err.IsServerErr {
			return emptyInput, departure, nil, entity_public.Product{}, entity_public.GetErrorToast("Internal error fetching departure", "")
		}
		return emptyInput, departure, nil, entity_public.Product{}, entity_public.GetWarningToast(err.Message, "")
	}

	// Validate departure type for NF-e: only "Self -> Third Party" qualifies
	if departure.Origin != nil {
		return emptyInput, departure, nil, entity_public.Product{}, entity_public.GetWarningToast(
			"Esta saída não pode emitir NF-e",
			"NF-e só pode ser emitida para saídas do tipo Próprio -> Terceiro (sem origem externa)")
	}
	if departure.Recipient == nil {
		return emptyInput, departure, nil, entity_public.Product{}, entity_public.GetWarningToast(
			"Esta saída não pode emitir NF-e",
			"Selecione um destinatário para emitir a NF-e")
	}

	// Get NFe farm config
	nfeModel := nfe_model.GetNFeModel()
	farmNFeConfig, dbErr := nfeModel.GetFarmConfig(farmID)
	if dbErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("GetFarmConfig error: %v", dbErr.Error()))
		return emptyInput, departure, nil, entity_public.Product{}, entity_public.GetErrorToast("Failed to get NFe configuration", "")
	}
	if farmNFeConfig == nil {
		return emptyInput, departure, nil, entity_public.Product{}, entity_public.GetWarningToast("NF-e não configurada", "acesse NF-e em Configurações")
	}

	// Get full farm data for emitter address
	farmConfigModel := farm_config_model.GetFarmConfigModel()
	farmData, farmErr := farmConfigModel.GetFarmConfig(farmID)
	if farmErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("GetFarmConfig error: %v", farmErr.Error()))
		return emptyInput, departure, nil, entity_public.Product{}, entity_public.GetErrorToast("Failed to get farm data", "")
	}
	if farmData == nil {
		return emptyInput, departure, nil, entity_public.Product{}, entity_public.GetWarningToast("Farm not found", "")
	}

	// Get recipient
	var recipient entity.RecipientData
	var person entity_public.FullPerson
	if departure.Recipient != nil {
		pModel := person_model.GetPersonModel()
		var personErr *model_error.ModelError
		person, personErr = pModel.GetFullPersonById(*departure.Recipient)
		if personErr != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("GetFullPersonById error: %v", personErr.Error()))
			return emptyInput, departure, nil, entity_public.Product{}, entity_public.GetErrorToast("Failed to get recipient info", "")
		}
		recipient = s.mapFullPersonToRecipient(person, nfeModel)
	}

	// Get emitter
	emitter := s.mapFarmToEmitter(farmNFeConfig, farmData, nfeModel)

	// Get the product for the crop — NCM and name come from the product table
	prodModel := product_model.GetProductModel()
	product, productErr := prodModel.GetProductById(departure.Crop)
	if productErr != nil {
		// Non-fatal: fall back to NCMSoja and a generic name
		model_error.GetLoggerModel().Log(fmt.Sprintf("GetProductById error: %v", productErr.Error()))
	}
	if product.NCM == "" || product.NCM == "00000000" {
		product.NCM = defaults.NCMSoja
	}
	if product.Name == "" {
		product.Name = "Produto Agricola"
	}

	// Resolve CFOP and unit from the farm config (per-farm defaults)
	effectiveCFOP := cfop
	if effectiveCFOP == "" {
		effectiveCFOP = farmNFeConfig.DefaultCFOP
	}
	effectiveUnit := farmNFeConfig.DefaultUnit
	if effectiveUnit == "" {
		effectiveUnit = "KG"
	}

	// Validate required emitter fields
	validationErrors := s.validateEmitter(emitter, farmNFeConfig.Serie)
	if len(validationErrors) > 0 {
		return emptyInput, departure, nil, entity_public.Product{}, entity_public.GetWarningToast(
			"Configuração de NF-e incompleta",
			strings.Join(validationErrors, "; "),
		)
	}

	// Validate required recipient fields
	recipientErrors := validateRecipient(recipient)
	if len(recipientErrors) > 0 {
		return emptyInput, departure, nil, entity_public.Product{}, entity_public.GetWarningToast(
			"Dados do destinatário incompletos",
			strings.Join(recipientErrors, "; "),
		)
	}

	// Merge user-typed rates with the farm config defaults. The result feeds
	// buildItems directly; the raw user input is returned separately so the
	// caller can persist a partial-override row in nfe_invoice_tax_rates.
	rates := MergeRates(userRates, farmNFeConfig)
	items := s.buildItems(departure, unitPrice, product.NCM, effectiveUnit, product.Name, farmNFeConfig, effectiveCFOP, rates)

	// Determine NaturezaOp from config or derive from CFOP
	naturezaOp := ""
	if farmNFeConfig.DefaultNaturezaOp != nil && *farmNFeConfig.DefaultNaturezaOp != "" {
		naturezaOp = *farmNFeConfig.DefaultNaturezaOp
	} else {
		naturezaOp = defaults.NaturezaOpForCFOP(effectiveCFOP)
	}

	// Build transport data: map departure vehicle + weights
	transport := entity.TransportData{
		ModFrete: farmNFeConfig.DefaultModFrete,
	}
	if departure.Vehicle > 0 {
		vModel := vehicle_model.GetVehicleModel()
		vehicle, vehErr := vModel.GetVehicle(departure.Vehicle)
		if vehErr == nil && vehicle.Plate != "" {
			transport.Veiculo = &entity.VeiculoData{
				Placa: vehicle.Plate,
				UF:    farmNFeConfig.EmitterUF,
			}
		}
	}
	if !departure.NetWeight.IsZero() || !departure.GrossWeight.IsZero() {
		transport.Volumes = []entity.VolumeData{{
			QVol:  1,
			Esp:   "Granel",
			PesoL: departure.NetWeight,
			PesoB: departure.GrossWeight,
		}}
	}

	// Build CND additional info: certificate blocks first, then all metadata.
	var cndParts []string
	var metaParts []string
	if farmBlock := formatCNDBlock("Fazenda", farmNFeConfig.CertificateNumber, farmNFeConfig.ExpDate); farmBlock != "" {
		cndParts = append(cndParts, farmBlock)
	}
	if personBlock := formatCNDBlock("Destinatário", person.CertificateNumber, person.ExpDate); personBlock != "" {
		cndParts = append(cndParts, personBlock)
	}
	if m := formatCNDMeta(farmNFeConfig.Meta); m != "" {
		metaParts = append(metaParts, m)
	}
	if m := formatCNDMeta(person.Meta); m != "" {
		metaParts = append(metaParts, m)
	}
	var infCplParts []string
	if len(cndParts) > 0 {
		infCplParts = append(infCplParts, strings.Join(cndParts, "; "))
	}
	if len(metaParts) > 0 {
		infCplParts = append(infCplParts, strings.Join(metaParts, "; "))
	}
	infCpl := strings.Join(infCplParts, "; ")

	// Build input
	input := entity.InvoiceInput{
		Serie:       farmNFeConfig.Serie,
		Numero:      0,
		Environment: farmNFeConfig.Environment,
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
					VPag:   departure.NetWeight.Mul(unitPrice),
				},
			},
		},
		TotalValue:            departure.NetWeight.Mul(unitPrice),
		InformacoesAdicionais: infCpl,
	}

	return input, departure, farmNFeConfig, product, entity_public.Toast{}
}

// BuildInvoiceFromDeparture builds, signs, and attempts to send an NF-e for a departure.
// If the normal SEFAZ is unavailable, it automatically tries SVC. If both are down,
// the invoice is saved as 'draft' and an error is returned — no DANFE is issued.
func (s *NFeService) BuildInvoiceFromDeparture(departureID uint32, unitPrice decimal.Decimal, farmID uint32, cfop string, userRates entity.TaxRates) (string, entity_public.Toast) {
	input, departure, farmNFeConfig, product, toast := s.prepareInvoiceBuildData(departureID, unitPrice, farmID, cfop, userRates)
	if toast.Type == entity_public.ErrorToast || toast.Type == entity_public.WarningToast {
		return "", toast
	}

	nfeModel := nfe_model.GetNFeModel()
	certPassword := decryptPassword(farmNFeConfig.CertificatePasswordEncrypted)
	sefazCfg := config.SefazConfig{
		Environment: config.Environment(farmNFeConfig.Environment),
		StateUF:     farmNFeConfig.EmitterUF,
		Timeout:     30 * time.Second,
	}
	invService := service.NewInvoiceService(sefazCfg)

	// --- Step 1: Normal emission attempt ---
	number, allocErr := nfeModel.AllocateNumber(farmID, farmNFeConfig.Serie)
	if allocErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("AllocateNumber error: %v", allocErr.Error()))
		return "", entity_public.GetErrorToast("Failed to allocate invoice number", "")
	}
	input.Numero = number
	input.CNF = generateRandomCNF()
	input.TpEmis = defaults.EmissaoNormal

	accessKey := entity.GenerateAccessKey(entity.AccessKeyData{
		CUF:      defaults.UFCode(input.Emitter.UF),
		AAMM:     time.Now().Format("0601"),
		Document: s.documentForAccessKey(input.Emitter),
		Mod:      defaults.ModeloNFe,
		Serie:    input.Serie,
		NNF:      input.Numero,
		TpEmis:   defaults.EmissaoNormal.String(),
		CNF:      input.CNF,
	})

	signedXML, signErr := invService.BuildAndSign(input, farmNFeConfig.CertificateData, certPassword)
	if signErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("BuildAndSign error: %v", signErr.Error()))
		return "", entity_public.GetErrorToast("Failed to build and sign NF-e", signErr.Error())
	}

	// Save initial record as draft (will be updated based on SEFAZ response)
	// Persist the raw user-typed rates (not the merged ones) so the row reflects
	// the user's exact intent and the worker reuses it on retry.
	var ratesToPersist *entity.TaxRates
	if hasAnyUserRate(userRates) {
		ratesToPersist = &userRates
	}
	invoiceID, createErr := nfeModel.CreateInvoice(departureID, accessKey, farmNFeConfig.Serie, number,
		farmNFeConfig.DefaultCFOP, product.NCM, departure.NetWeight, unitPrice, input.TotalValue,
		ratesToPersist)
	if createErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("CreateInvoice error: %v", createErr.Error()))
		return "", entity_public.GetErrorToast("Failed to save invoice", "")
	}

	// Persist the signed XML regardless of the SEFAZ outcome: the XML download
	// falls back to it, and the retry worker builds the <nfeProc> from it when
	// a pending invoice later gets authorized. The status stays 'draft' until
	// a real SEFAZ response arrives.
	if xmlErr := nfeModel.UpdateInvoiceSignedXML(invoiceID, signedXML); xmlErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateInvoiceSignedXML error: %v", xmlErr.Error()))
	}

	// Attempt synchronous normal SEFAZ submission
	sefazResp, sendErr := invService.SendToSefaz(signedXML, farmNFeConfig.CertificateData, certPassword)
	if sendErr == nil && sefazResp != nil {
		// We got a response — handle it normally
		return s.handleSefazResponse(sefazResp, invoiceID, signedXML, nfeModel)
	}

	// --- Step 2: Normal SEFAZ failed (network error or no response) ---
	// Check if SVC is operational
	svcResp, svcErr := invService.CheckSVCStatus(farmNFeConfig.CertificateData, certPassword)
	if svcErr == nil && svcResp != nil && svcResp.IsSVCOperational() {
		// SVC is active — rebuild with new number and contingency settings
		return s.attemptSVCContingency(input, departure, farmNFeConfig, product, signedXML, invoiceID, accessKey, nfeModel, invService, certPassword, unitPrice, ratesToPersist)
	}

	// --- Step 3: Both SEFAZ and SVC are unavailable ---
	// Keep the invoice as draft. Do NOT allow DANFE generation.
	if sendErr != nil {
		errUpd := nfeModel.UpdateInvoiceStatus(invoiceID, "draft", "", "", "SEFAZ e SVC indisponiveis: "+sendErr.Error())
		if errUpd != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateInvoiceStatus error: %v", errUpd.Error()))
		}
	} else {
		errUpd := nfeModel.UpdateInvoiceStatus(invoiceID, "draft", "", "", "SEFAZ e SVC indisponiveis: resposta vazia")
		if errUpd != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateInvoiceStatus error: %v", errUpd.Error()))
		}
	}
	return "", entity_public.GetErrorToast(
		"SEFAZ e SVC indisponiveis",
		"NF-e nao pode ser emitida no momento. Tente novamente mais tarde.",
	)
}

// handleSefazResponse processes the SEFAZ response and updates the invoice status.
func (s *NFeService) handleSefazResponse(sefazResp *sefaz.AutorizacaoResponse, invoiceID int, signedXML string, nfeModel *nfe_model.NFeModel) (string, entity_public.Toast) {
	if sefazResp.IsAuthorized() {
		errUpd := nfeModel.UpdateInvoiceStatus(invoiceID, "authorized", sefazResp.Protocol, sefazResp.StatusCode, sefazResp.StatusMotive)
		if errUpd != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateInvoiceStatus error: %v", errUpd.Error()))
		}
		// Build and store the <nfeProc> wrapper (signed NFe + protocol) so the
		// DANFE parser can extract nProt/dhRecbto from the stored XML.
		if authXML, buildErr := nfe_xml.BuildAuthorizedXML(signedXML, sefazResp.AccessKey, sefazResp.Protocol, sefazResp.DhRecbto, sefazResp.StatusCode, sefazResp.StatusMotive); buildErr == nil {
			if xmlErr := nfeModel.UpdateInvoiceAuthorizedXML(invoiceID, authXML); xmlErr != nil {
				model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateInvoiceAuthorizedXML error: %v", xmlErr.Error()))
			}
		} else {
			model_error.GetLoggerModel().Log(fmt.Sprintf("BuildAuthorizedXML error: %v", buildErr.Error()))
		}
		return signedXML, entity_public.GetSuccessToast("NF-e autorizada pela SEFAZ", fmt.Sprintf("Protocolo: %s", sefazResp.Protocol))
	}

	if sefazResp.IsProcessing() || sefazResp.IsAccepted() {
		errUpd := nfeModel.UpdateInvoiceStatus(invoiceID, "pending", sefazResp.Protocol, sefazResp.StatusCode, sefazResp.StatusMotive)
		if errUpd != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateInvoiceStatus error: %v", errUpd.Error()))
		}
		return signedXML, entity_public.GetSuccessToast("NF-e enviada à SEFAZ e em processamento", fmt.Sprintf("Status: %s - %s", sefazResp.StatusCode, sefazResp.StatusMotive))
	}

	if sefazResp.IsRejected() {
		errUpd := nfeModel.UpdateInvoiceStatus(invoiceID, "denied", "", sefazResp.StatusCode, sefazResp.StatusMotive)
		if errUpd != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateInvoiceStatus error: %v", errUpd.Error()))
		}
		return signedXML, entity_public.GetErrorToast(
			fmt.Sprintf("NF-e rejeitada pela SEFAZ (%s)", sefazResp.StatusCode),
			sefazResp.StatusMotive,
		)
	}

	// Unknown status — queue for status polling
	errUpd := nfeModel.UpdateInvoiceStatus(invoiceID, "pending", "", sefazResp.StatusCode, sefazResp.StatusMotive)
	if errUpd != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateInvoiceStatus error: %v", errUpd.Error()))
	}
	return signedXML, entity_public.GetSuccessToast("NF-e enviada à SEFAZ e em processamento", fmt.Sprintf("Status: %s", sefazResp.StatusMotive))
}

// attemptSVCContingency tries to send the NF-e via SVC with a new number and contingency fields.
func (s *NFeService) attemptSVCContingency(input entity.InvoiceInput, departure entity_public.Departure, farmNFeConfig *entity_public.FarmConfig, product entity_public.Product, oldSignedXML string, oldInvoiceID int, oldAccessKey string, nfeModel *nfe_model.NFeModel, invService *service.InvoiceService, certPassword string, unitPrice decimal.Decimal, taxRates *entity.TaxRates) (string, entity_public.Toast) {
	// Allocate new number for the contingency invoice (required by MOC to avoid duplicate Chave Natural)
	newNumber, allocErr := nfeModel.AllocateNumber(uint32(departure.Farm), farmNFeConfig.Serie)
	if allocErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("AllocateNumber (SVC) error: %v", allocErr.Error()))
		return "", entity_public.GetErrorToast("Failed to allocate contingency invoice number", "")
	}

	tpEmis := defaults.SVCForState(farmNFeConfig.EmitterUF)
	now := time.Now()
	reason := "Indisponibilidade do ambiente de autorizacao da SEFAZ de origem"

	newSignedXML, newAccessKey, rebuildErr := invService.RebuildForContingency(input, newNumber, generateRandomCNF(), tpEmis, reason, farmNFeConfig.CertificateData, certPassword)
	if rebuildErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("RebuildForContingency error: %v", rebuildErr.Error()))
		return "", entity_public.GetErrorToast("Failed to rebuild NF-e for SVC contingency", rebuildErr.Error())
	}

	// Save the new contingency invoice (same tax rates as the original attempt)
	newInvoiceID, createErr := nfeModel.CreateInvoiceWithEmission(
		departure.Id, newAccessKey, farmNFeConfig.Serie, newNumber,
		farmNFeConfig.DefaultCFOP, product.NCM, departure.NetWeight, unitPrice, input.TotalValue,
		int(tpEmis), now, reason, taxRates,
	)
	if createErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("CreateInvoice (SVC) error: %v", createErr.Error()))
		return "", entity_public.GetErrorToast("Failed to save contingency invoice", "")
	}

	xmlErr := nfeModel.UpdateInvoiceSignedXML(newInvoiceID, newSignedXML)
	if xmlErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateInvoiceSignedXML (SVC) error: %v", xmlErr.Error()))
		return "", entity_public.GetErrorToast("Failed to save contingency signed XML", "")
	}

	// Send to SVC
	sefazResp, sendErr := invService.SendToSefazWithEmission(newSignedXML, farmNFeConfig.CertificateData, certPassword, tpEmis)
	if sendErr != nil {
		// SVC also failed — mark new invoice as draft, supersede old one
		errUpd := nfeModel.UpdateInvoiceStatus(newInvoiceID, "draft", "", "", "SVC indisponivel: "+sendErr.Error())
		if errUpd != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateInvoiceStatus (SVC) error: %v", errUpd.Error()))
		}
		_ = nfeModel.SupersedeInvoice(oldInvoiceID, newInvoiceID)
		return "", entity_public.GetErrorToast(
			"SEFAZ e SVC indisponiveis",
			"NF-e nao pode ser emitida no momento. Tente novamente mais tarde.",
		)
	}

	if sefazResp == nil {
		errUpd := nfeModel.UpdateInvoiceStatus(newInvoiceID, "draft", "", "", "SVC indisponivel: resposta vazia")
		if errUpd != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateInvoiceStatus (SVC) error: %v", errUpd.Error()))
		}
		_ = nfeModel.SupersedeInvoice(oldInvoiceID, newInvoiceID)
		return "", entity_public.GetErrorToast(
			"SEFAZ e SVC indisponiveis",
			"NF-e nao pode ser emitida no momento. Tente novamente mais tarde.",
		)
	}

	// SVC responded — handle normally and supersede the old normal attempt
	resultXML, resultToast := s.handleSefazResponse(sefazResp, newInvoiceID, newSignedXML, nfeModel)
	_ = nfeModel.SupersedeInvoice(oldInvoiceID, newInvoiceID)
	return resultXML, resultToast
}

// CancelInvoice registers a cancellation event (110111) for an authorized NF-e
// at the same environment that authorized it (origin SEFAZ or SVC). On success
// the invoice is marked as 'cancelled' and the signed event XML is stored as
// legal proof.
func (s *NFeService) CancelInvoice(accessKey, justification string, farmID uint32) entity_public.Toast {
	nfeModel := nfe_model.GetNFeModel()
	invoice, err := nfeModel.GetInvoiceByAccessKey(accessKey)
	if err != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("CancelInvoice GetInvoiceByAccessKey error: %v", err.Error()))
		return entity_public.GetErrorToast("Erro interno ao buscar a NF-e", "")
	}
	if invoice == nil {
		return entity_public.GetWarningToast("NF-e não encontrada", "")
	}

	// Verify the invoice belongs to the user's farm
	dModel := departure_model.GetDepartureModel()
	departure, depErr := dModel.GetDeparture(invoice.DepartureID)
	if depErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("CancelInvoice GetDeparture error: %v", depErr.Error()))
		return entity_public.GetErrorToast("Erro interno ao validar a NF-e", "")
	}
	if departure.Farm != farmID {
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
		model_error.GetLoggerModel().Log(fmt.Sprintf("CancelInvoice GetFarmConfig error: %v", dbErr.Error()))
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

	// Per the MOC contingency annex, the cancellation must be registered in the
	// same environment that authorized the NF-e (origin SEFAZ or SVC).
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
		updErr := nfeModel.UpdateInvoiceCancelled(invoice.ID, justification, signedEventXML, resp.StatusCode, resp.StatusMotive)
		if updErr != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateInvoiceCancelled error: %v", updErr.Error()))
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

	return entity_public.GetWarningToast(
		fmt.Sprintf("SEFAZ rejeitou o cancelamento (%s)", resp.StatusCode),
		resp.StatusMotive,
	)
}

// GeneratePreviewDANFE builds a preview DANFE PDF without allocating an invoice number,
// signing, or persisting anything. It returns the PDF bytes and any validation toast.
func (s *NFeService) GeneratePreviewDANFE(departureID uint32, unitPrice decimal.Decimal, farmID uint32, cfop string, userRates entity.TaxRates) ([]byte, entity_public.Toast) {
	input, _, _, _, toast := s.prepareInvoiceBuildData(departureID, unitPrice, farmID, cfop, userRates)
	if toast.Type == entity_public.ErrorToast || toast.Type == entity_public.WarningToast {
		return nil, toast
	}

	now := time.Now()
	product := input.Items[0].Produto

	item := input.Items[0]
	imp := item.Imposto
	transport := input.Transport

	tpAmb := "1"
	if input.Environment == 2 {
		tpAmb = "2"
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
		Products: []entity.DANFEProduct{
			{
				Code:      product.Codigo,
				Desc:      product.XProd,
				NCM:       product.NCM,
				CST:       imp.ICMS.CST,
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
			},
		},
		TotalValue: input.TotalValue,
		VBC:        imp.ICMS.VBC,
		VICMS:      imp.ICMS.VICMS,
		VPIS:       imp.PIS.VPIS,
		VCOFINS:    imp.COFINS.VCOFINS,
		ModFrete:   strconv.Itoa(transport.ModFrete),
		InfCpl:     input.InformacoesAdicionais,
	}

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

	generator := service.NewDANFEGenerator()
	pdfBytes, genErr := generator.GeneratePreview(data)
	if genErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("GeneratePreview error: %v", genErr.Error()))
		return nil, entity_public.GetErrorToast("Failed to generate preview DANFE", genErr.Error())
	}

	return pdfBytes, entity_public.Toast{}
}

// formatDocument returns the non-empty document (CNPJ or CPF) for display.
func (s *NFeService) formatDocument(cnpj, cpf string) string {
	if cnpj != "" {
		return cnpj
	}
	return cpf
}

// documentForAccessKey returns the CNPJ or zero-padded CPF for the access key.
func (s *NFeService) documentForAccessKey(emitter entity.EmitterData) string {
	if emitter.Type == 2 {
		// CPF: pad with leading zeros to 14 digits
		return padLeftZeros(emitter.Document, 14)
	}
	// CNPJ must be exactly 14 digits
	return padLeftZeros(emitter.Document, 14)
}

func padLeftZeros(s string, length int) string {
	for len(s) < length {
		s = "0" + s
	}
	if len(s) > length {
		return s[:length]
	}
	return s
}

// buildItems creates the NF-e item data from departure and configs.
func (s *NFeService) buildItems(departure entity_public.Departure, unitPrice decimal.Decimal, ncm string, unit string, description string, farmConfig *entity_public.FarmConfig, cfop string, rates entity.TaxRates) []entity.ItemData {
	totalValue := departure.NetWeight.Mul(unitPrice)
	regime := defaults.TaxRegime(farmConfig.TaxRegime)

	// Resolved rates are authoritative (mergeRates already applied user > farm > last-resort).
	icmsRate := *rates.ICMSRate
	pisRate := *rates.PISRate
	cofinsRate := *rates.COFINSRate

	// Determine ICMS CST/CSOSN from farm config or defaults
	var icmsCST, icmsCSOSN string
	var vBC, vICMS decimal.Decimal
	var picms decimal.Decimal

	if regime == defaults.TaxRegimeSimplesNacional {
		if farmConfig.DefaultICMSCST != nil && *farmConfig.DefaultICMSCST != "" {
			icmsCSOSN = *farmConfig.DefaultICMSCST
		} else {
			icmsCSOSN = defaults.CSOSNSemPermissaoCredito // 102
		}
		// For SN, ICMS totals are zero; pICMS is not sent in CSOSN groups
		vBC = decimal.Zero
		vICMS = decimal.Zero
		picms = decimal.Zero
	} else {
		if farmConfig.DefaultICMSCST != nil && *farmConfig.DefaultICMSCST != "" {
			icmsCST = *farmConfig.DefaultICMSCST
		} else {
			icmsCST = defaults.CSTTributadaIntegral // 00
		}
		vBC = totalValue
		vICMS = totalValue.Mul(icmsRate)
		picms = icmsRate.Mul(decimal.NewFromInt(100)) // rate as percentage
	}

	// Determine PIS and COFINS CST from farm config
	pisCST := "01"
	if farmConfig.DefaultPISCST != nil && *farmConfig.DefaultPISCST != "" {
		pisCST = *farmConfig.DefaultPISCST
	}
	cofinsCST := "01"
	if farmConfig.DefaultCOFINSCST != nil && *farmConfig.DefaultCOFINSCST != "" {
		cofinsCST = *farmConfig.DefaultCOFINSCST
	}

	// Description fallback: if empty, try the product defaults; finally "Produto Agricola"
	if description == "" {
		productDefaults := defaults.GetProductDefaults("", regime)
		if productDefaults.Description != "" {
			description = productDefaults.Description
		} else {
			description = "Produto Agricola"
		}
	}

	return []entity.ItemData{
		{
			Numero: 1,
			Produto: entity.ProdutoData{
				Codigo:   strconv.Itoa(int(departure.Crop)),
				CEAN:     "SEM GTIN",
				XProd:    description,
				NCM:      ncm,
				CFOP:     cfop,
				UCom:     unit,
				QCom:     departure.NetWeight,
				VUnCom:   unitPrice,
				VProd:    totalValue,
				CEANTrib: "SEM GTIN",
				UTrib:    unit,
				QTrib:    departure.NetWeight,
				VUnTrib:  unitPrice,
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
					VBC:  totalValue,
					PPIS: pisRate.Mul(decimal.NewFromInt(100)),
					VPIS: totalValue.Mul(pisRate),
				},
				COFINS: entity.COFINSData{
					CST:     cofinsCST,
					VBC:     totalValue,
					PCOFINS: cofinsRate.Mul(decimal.NewFromInt(100)),
					VCOFINS: totalValue.Mul(cofinsRate),
				},
			},
		},
	}
}

// MergeRates resolves the effective TaxRates for an emission.
// User-provided rates are authoritative when non-nil (even if zero). For nil
// values, fall back to the farm config; if the farm config is also zero for
// that rate, use the last-resort default (the historical pre-migration values
// for grain sales).
//
// All three returned pointers are non-nil by construction, so callers (e.g.
// buildItems) can safely dereference without nil checks.
func MergeRates(overrides entity.TaxRates, farmConfig *entity_public.FarmConfig) entity.TaxRates {
	var pICMS, pPIS, pCOFINS decimal.Decimal
	if farmConfig != nil {
		pICMS, pPIS, pCOFINS = farmConfig.ICMSRate, farmConfig.PISRate, farmConfig.COFINSRate
	}
	return entity.TaxRates{
		ICMSRate:   pickRate(overrides.ICMSRate, pICMS, decimal.NewFromFloat(0.17)),
		PISRate:    pickRate(overrides.PISRate, pPIS, decimal.NewFromFloat(0.0165)),
		COFINSRate: pickRate(overrides.COFINSRate, pCOFINS, decimal.NewFromFloat(0.076)),
	}
}

func pickRate(override *decimal.Decimal, product, fallback decimal.Decimal) *decimal.Decimal {
	if override != nil {
		return override
	}
	if !product.IsZero() {
		return &product
	}
	return &fallback
}

// hasAnyUserRate reports whether the user typed at least one rate in the form.
// Used to decide whether to persist a row in nfe_invoice_tax_rates.
func hasAnyUserRate(rates entity.TaxRates) bool {
	return rates.ICMSRate != nil || rates.PISRate != nil || rates.COFINSRate != nil
}

// mapFarmToEmitter maps farm config and farm data to NF-e emitter data.
func (s *NFeService) mapFarmToEmitter(cfg *entity_public.FarmConfig, farm *entity_public.Farm, nfeModel *nfe_model.NFeModel) entity.EmitterData {
	emitter := entity.EmitterData{
		Type:       cfg.EmitterType,
		IE:         cfg.IEEmitter,
		XNome:      "",
		Logradouro: "",
		Numero:     "",
		Bairro:     "",
		CodigoMun:  "",
		Municipio:  "",
		UF:         cfg.EmitterUF,
		CEP:        "",
		Fone:       "",
		CRT:        defaults.TaxRegime(cfg.TaxRegime).CRT(),
		Document:   *cfg.DocEmitter,
	}

	// Populate emitter name and address from farm data
	if farm.Name != nil {
		emitter.XNome = *farm.Name
	}
	if farm.StorageName != nil {
		emitter.XFant = *farm.StorageName
	}
	if farm.Street != nil {
		emitter.Logradouro = *farm.Street
	}
	if farm.Number != nil {
		emitter.Numero = strconv.FormatUint(uint64(*farm.Number), 10)
	}
	if farm.Neighborhood != nil {
		emitter.Bairro = *farm.Neighborhood
	}
	if farm.City != nil {
		emitter.Municipio = *farm.City
	}
	if farm.State != nil {
		emitter.UF = *farm.State
	}
	if farm.Cep != nil {
		emitter.CEP = *farm.Cep
	}
	if farm.PhoneNumber != nil {
		emitter.Fone = *farm.PhoneNumber
	}

	// Resolve IBGE municipality code
	if farm.City != nil && farm.State != nil {
		code, _ := nfeModel.GetMunicipio(*farm.City, *farm.State)
		if code != "" {
			emitter.CodigoMun = code
		}
	}

	return emitter
}

// validateEmitter checks that all required emitter fields are populated.
func (s *NFeService) validateEmitter(emit entity.EmitterData, serie int) []string {
	var errs []string

	if emit.Document == "" {
		if emit.Type == 2 {
			errs = append(errs, "CPF do emitente não configurado")
		} else {
			errs = append(errs, "CNPJ do emitente não configurado")
		}
	}

	if emit.Type == 2 {
		// CPF (Produtor Rural) emitters must use series 920-969 per Nota Técnica 2018.001
		if serie < 920 || serie > 969 {
			errs = append(errs, fmt.Sprintf("Série %d inválida para emitente CPF: deve estar entre 920 e 969", serie))
		}
	} else {
		// CNPJ emitters must use series 0-889
		if serie < 0 || serie > 889 {
			errs = append(errs, fmt.Sprintf("Série %d inválida para emitente CNPJ: deve estar entre 0 e 889", serie))
		}
	}

	if emit.XNome == "" {
		errs = append(errs, "Nome/Razão Social do emitente não configurado")
	}
	if emit.Logradouro == "" {
		errs = append(errs, "Logradouro do emitente não configurado")
	}
	if emit.Numero == "" {
		errs = append(errs, "Número do endereço do emitente não configurado")
	}
	if emit.CodigoMun == "" {
		errs = append(errs, "Município do emitente não configurado (código IBGE)")
	}
	if emit.Municipio == "" {
		errs = append(errs, "Nome do município do emitente não configurado")
	}
	if emit.UF == "" {
		errs = append(errs, "UF do emitente não configurada")
	}
	if emit.CEP == "" {
		errs = append(errs, "CEP do emitente não configurado")
	}
	if emit.Bairro == "" {
		errs = append(errs, "Bairro do emitente não configurado")
	}
	return errs
}

// validateRecipient checks mandatory dest fields per MOC Anexo I (Grupo E).
func validateRecipient(dest entity.RecipientData) []string {
	var errs []string
	if dest.Type == 2 {
		if dest.CPF == "" {
			errs = append(errs, "CPF do destinatário não informado")
		}
	} else {
		if dest.CNPJ == "" {
			errs = append(errs, "CNPJ do destinatário não informado")
		}
	}
	if dest.XNome == "" {
		errs = append(errs, "Nome/Razão Social do destinatário não informado")
	}
	if dest.Logradouro == "" {
		errs = append(errs, "Logradouro do destinatário não informado")
	}
	if dest.Numero == "" {
		errs = append(errs, "Número do endereço do destinatário não informado")
	}
	if dest.Bairro == "" {
		errs = append(errs, "Bairro do destinatário não informado")
	}
	if dest.CodigoMun == "" {
		errs = append(errs, "Município do destinatário não informado (código IBGE)")
	}
	if dest.Municipio == "" {
		errs = append(errs, "Nome do município do destinatário não informado")
	}
	if dest.UF == "" {
		errs = append(errs, "UF do destinatário não informada")
	}
	return errs
}

// isValidIEMT validates an Inscrição Estadual for Mato Grosso.
func isValidIEMT(ie string) bool {
	if ie == "" {
		return false
	}
	cleaned := cleanIE(ie)
	return len(cleaned) == 9
}

// cleanIE strips any non-numeric characters from an IE string.
func cleanIE(ie string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, ie)
}

// mapFullPersonToRecipient maps a full person record to NF-e recipient data.
// NF-e Type convention: 1=CNPJ (legal person), 2=CPF (natural person)
// Our FullPerson.Type: 1=natural, 2=legal
func (s *NFeService) mapFullPersonToRecipient(person entity_public.FullPerson, nfeModel *nfe_model.NFeModel) entity.RecipientData {
	// Map our person type to NF-e type
	// person.Type 1 (natural) → NF-e Type 2 (CPF)
	// person.Type 2 (legal)   → NF-e Type 1 (CNPJ)
	var nfeType int
	if person.Type == 1 {
		nfeType = 2 // CPF
	} else {
		nfeType = 1 // CNPJ
	}

	recipient := entity.RecipientData{
		Type:       nfeType,
		XNome:      person.Name,
		Logradouro: "",
		Numero:     "",
		Bairro:     "",
		CodigoMun:  "",
		Municipio:  "",
		UF:         "",
		CEP:        "",
		Fone:       "",
		IndIEDest:  "9", // Default: non-contributor
	}

	if nfeType == 1 {
		recipient.CNPJ = person.Document
	} else {
		recipient.CPF = person.Document
	}

	// IE and indIEDest
	if isValidIEMT(person.IE) {
		recipient.IE = cleanIE(person.IE)
		recipient.IndIEDest = "1" // Contribuinte ICMS
	}

	// Address
	if person.Street != nil {
		recipient.Logradouro = *person.Street
	}
	if person.Number != nil {
		recipient.Numero = strconv.FormatUint(uint64(*person.Number), 10)
	}
	if recipient.Numero == "" {
		recipient.Numero = "S/N"
	}
	if person.Neighborhood != nil {
		recipient.Bairro = *person.Neighborhood
	}
	if person.City != nil {
		recipient.Municipio = *person.City
	}
	if person.State != nil {
		recipient.UF = *person.State
	}
	if person.Cep != nil {
		recipient.CEP = *person.Cep
	}
	if person.PhoneNumber != nil {
		recipient.Fone = *person.PhoneNumber
	}

	// Resolve IBGE municipality code
	if person.City != nil && person.State != nil {
		code, _ := nfeModel.GetMunicipio(*person.City, *person.State)
		if code != "" {
			recipient.CodigoMun = code
		}
	}

	return recipient
}

// formatCNDBlock returns the certificate portion of a CND block for the DANFE infCpl.
// If the certificate number is absent, it returns an empty string.
func formatCNDBlock(label string, certNumber *string, expDate *time.Time) string {
	if certNumber == nil || *certNumber == "" {
		return ""
	}
	var parts []string
	parts = append(parts, fmt.Sprintf("CND %s:", label))
	certParts := []string{fmt.Sprintf("Cert. Nº %s", *certNumber)}
	if expDate != nil {
		certParts = append(certParts, fmt.Sprintf("válido até %s", expDate.Format("02/01/2006")))
	}
	parts = append(parts, strings.Join(certParts, ", "))
	// Single space separator: the schema TString pattern forbids newlines in
	// infCpl (SEFAZ rejects with cvc-type.3.1.3 otherwise).
	return strings.Join(parts, " ")
}

// formatCNDMeta returns the metadata lines for a CND entry.
// If there is no metadata, it returns an empty string.
func formatCNDMeta(meta *map[string]interface{}) string {
	if meta == nil || len(*meta) == 0 {
		return ""
	}
	var parts []string
	for k, v := range *meta {
		parts = append(parts, fmt.Sprintf("%s: %v", k, v))
	}
	// "; " separator: the schema TString pattern forbids newlines in infCpl
	// (SEFAZ rejects with cvc-type.3.1.3 otherwise).
	return strings.Join(parts, "; ")
}

// EncryptPassword encrypts a certificate password using AES-256-GCM with the env key.
// Returns base64-encoded ciphertext (nonce || ciphertext || tag).
func EncryptPassword(plaintext string) string {
	key := os.Getenv("NFE_CERT_KEY")
	if key == "" {
		// In development without key, return plaintext
		return plaintext
	}
	return encryptAES(plaintext, key)
}

// decryptPassword decrypts the certificate password using the environment key.
func decryptPassword(encrypted string) string {
	key := os.Getenv("NFE_CERT_KEY")
	if key == "" {
		return encrypted // In development, may be plain text
	}
	return decryptAES(encrypted, key)
}

func encryptAES(plaintext, keyStr string) string {
	key := deriveKey(keyStr)

	block, err := aes.NewCipher(key)
	if err != nil {
		return plaintext
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return plaintext
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return plaintext
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func decryptAES(ciphertextB64, keyStr string) string {
	key := deriveKey(keyStr)

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return ciphertextB64
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return ciphertextB64
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return ciphertextB64
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return ciphertextB64
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return ciphertextB64
	}

	return string(plaintext)
}

func deriveKey(keyStr string) []byte {
	// Derive a 32-byte key from the provided string using SHA-256-like hex truncation
	// For simplicity, we hash the string to hex and take first 32 bytes, or pad/repeat
	keyBytes := []byte(keyStr)
	if len(keyBytes) >= 32 {
		return keyBytes[:32]
	}
	// Pad by repeating
	padded := make([]byte, 32)
	for i := range 32 {
		padded[i] = keyBytes[i%len(keyBytes)]
	}
	return padded
}

// generateRandomCNF generates a random 8-digit numeric code for the NF-e cNF field.
// SEFAZ requires this to be a random value, not sequential or derived from nNF.
func generateRandomCNF() string {
	// Range: 10000000 to 99999999 (8 digits, no leading zeros)
	max := big.NewInt(90000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		// Fallback: use current nanoseconds modulo range
		return fmt.Sprintf("%08d", (time.Now().UnixNano()%90000000)+10000000)
	}
	return fmt.Sprintf("%08d", n.Int64()+10000000)
}
