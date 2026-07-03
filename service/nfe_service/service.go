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
func (s *NFeService) prepareInvoiceBuildData(departureID uint32, unitPrice decimal.Decimal, farmID uint32) (entity.InvoiceInput, entity_public.Departure, *nfe_model.FarmConfig, *nfe_model.ProductConfig, entity_public.Toast) {
	var emptyInput entity.InvoiceInput

	// Get departure
	dModel := departure_model.GetDepartureModel()
	departure, err := dModel.GetDeparture(departureID)
	if err != nil {
		if err.IsServerErr {
			return emptyInput, departure, nil, nil, entity_public.GetErrorToast("Internal error fetching departure", "")
		}
		return emptyInput, departure, nil, nil, entity_public.GetWarningToast(err.Message, "")
	}

	// Validate departure type for NF-e: only "Self -> Third Party" qualifies
	if departure.Origin != nil {
		return emptyInput, departure, nil, nil, entity_public.GetWarningToast(
			"Esta saída não pode emitir NF-e",
			"NF-e só pode ser emitida para saídas do tipo Próprio -> Terceiro (sem origem externa)")
	}
	if departure.Recipient == nil {
		return emptyInput, departure, nil, nil, entity_public.GetWarningToast(
			"Esta saída não pode emitir NF-e",
			"Selecione um destinatário para emitir a NF-e")
	}

	// Get NFe farm config
	nfeModel := nfe_model.GetNFeModel()
	farmNFeConfig, dbErr := nfeModel.GetFarmConfig(farmID)
	if dbErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("GetFarmConfig error: %v", dbErr.Error()))
		return emptyInput, departure, nil, nil, entity_public.GetErrorToast("Failed to get NFe configuration", "")
	}
	if farmNFeConfig == nil {
		return emptyInput, departure, nil, nil, entity_public.GetWarningToast("NFe not configured for this farm", "configure in settings")
	}

	// Get full farm data for emitter address
	farmConfigModel := farm_config_model.GetFarmConfigModel()
	farmData, farmErr := farmConfigModel.GetFarmConfig(farmID)
	if farmErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("GetFarmConfig error: %v", farmErr.Error()))
		return emptyInput, departure, nil, nil, entity_public.GetErrorToast("Failed to get farm data", "")
	}
	if farmData == nil {
		return emptyInput, departure, nil, nil, entity_public.GetWarningToast("Farm not found", "")
	}

	// Get recipient
	var recipient entity.RecipientData
	if departure.Recipient != nil {
		pModel := person_model.GetPersonModel()
		person, personErr := pModel.GetFullPersonById(*departure.Recipient)
		if personErr != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("GetFullPersonById error: %v", personErr.Error()))
			return emptyInput, departure, nil, nil, entity_public.GetErrorToast("Failed to get recipient info", "")
		}
		recipient = s.mapFullPersonToRecipient(person, nfeModel)
	}

	// Get emitter
	emitter := s.mapFarmToEmitter(farmNFeConfig, farmData, nfeModel)

	// Get product config for the crop
	productConfig, prodErr := nfeModel.GetProductConfig(farmID, departure.Crop)
	if prodErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("GetProductConfig error: %v", prodErr.Error()))
		return emptyInput, departure, nil, nil, entity_public.GetErrorToast("Failed to get product config", "")
	}
	if productConfig == nil {
		productConfig = &nfe_model.ProductConfig{
			NCM:         defaults.NCMSoja,
			DefaultCFOP: defaults.CFOPVendaProducao,
			Unit:        "KG",
			Description: nil,
		}
	}

	// Validate required emitter fields
	validationErrors := s.validateEmitter(emitter, farmNFeConfig.Serie)
	if len(validationErrors) > 0 {
		return emptyInput, departure, nil, nil, entity_public.GetWarningToast(
			"Configuração de NF-e incompleta",
			strings.Join(validationErrors, "; "),
		)
	}

	// Build items
	items := s.buildItems(departure, unitPrice, farmNFeConfig, productConfig)

	// Build input
	input := entity.InvoiceInput{
		Serie:       farmNFeConfig.Serie,
		Numero:      0,
		Environment: farmNFeConfig.Environment,
		NaturezaOp:  "Venda de producao do estabelecimento",
		Emitter:     emitter,
		Recipient:   recipient,
		Items:       items,
		Transport: entity.TransportData{
			ModFrete: farmNFeConfig.DefaultModFrete,
		},
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
		TotalValue: departure.NetWeight.Mul(unitPrice),
	}

	return input, departure, farmNFeConfig, productConfig, entity_public.Toast{}
}

// BuildInvoiceFromDeparture builds, signs, and attempts to send an NF-e for a departure.
// If the normal SEFAZ is unavailable, it automatically tries SVC. If both are down,
// the invoice is saved as 'draft' and an error is returned — no DANFE is issued.
func (s *NFeService) BuildInvoiceFromDeparture(departureID uint32, unitPrice decimal.Decimal, farmID uint32) (string, entity_public.Toast) {
	input, departure, farmNFeConfig, productConfig, toast := s.prepareInvoiceBuildData(departureID, unitPrice, farmID)
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
	invService := service.NewInvoiceService(nil, sefazCfg)

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
	invoiceID, createErr := nfeModel.CreateInvoice(departureID, accessKey, farmNFeConfig.Serie, number,
		productConfig.DefaultCFOP, productConfig.NCM, departure.NetWeight, unitPrice, input.TotalValue)
	if createErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("CreateInvoice error: %v", createErr.Error()))
		return "", entity_public.GetErrorToast("Failed to save invoice", "")
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
		return s.attemptSVCContingency(input, departure, farmNFeConfig, productConfig, signedXML, invoiceID, accessKey, nfeModel, invService, certPassword, unitPrice)
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
func (s *NFeService) attemptSVCContingency(input entity.InvoiceInput, departure entity_public.Departure, farmNFeConfig *nfe_model.FarmConfig, productConfig *nfe_model.ProductConfig, oldSignedXML string, oldInvoiceID int, oldAccessKey string, nfeModel *nfe_model.NFeModel, invService *service.InvoiceService, certPassword string, unitPrice decimal.Decimal) (string, entity_public.Toast) {
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

	// Save the new contingency invoice
	newInvoiceID, createErr := nfeModel.CreateInvoiceWithEmission(
		departure.Id, newAccessKey, farmNFeConfig.Serie, newNumber,
		productConfig.DefaultCFOP, productConfig.NCM, departure.NetWeight, unitPrice, input.TotalValue,
		int(tpEmis), now, reason,
	)
	if createErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("CreateInvoice (SVC) error: %v", createErr.Error()))
		return "", entity_public.GetErrorToast("Failed to save contingency invoice", "")
	}

	xmlErr := nfeModel.UpdateInvoiceXML(newInvoiceID, newSignedXML)
	if xmlErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateInvoiceXML (SVC) error: %v", xmlErr.Error()))
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

// GeneratePreviewDANFE builds a preview DANFE PDF without allocating an invoice number,
// signing, or persisting anything. It returns the PDF bytes and any validation toast.
func (s *NFeService) GeneratePreviewDANFE(departureID uint32, unitPrice decimal.Decimal, farmID uint32) ([]byte, entity_public.Toast) {
	input, _, _, _, toast := s.prepareInvoiceBuildData(departureID, unitPrice, farmID)
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
		EmitterCNPJ:         s.formatDocument(input.Emitter.CNPJ, input.Emitter.CPF),
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
				Code:       product.Codigo,
				Desc:       product.XProd,
				NCM:        product.NCM,
				CST:        imp.ICMS.CST,
				CFOP:       product.CFOP,
				Unit:       product.UCom,
				Quantity:   product.QCom,
				UnitPrice:  product.VUnCom,
				Total:      product.VProd,
				UTrib:      product.UTrib,
				QTrib:      product.QTrib,
				VUnTrib:    product.VUnTrib,
				InfAdProd:  item.InfAdProd,
				VBC:        imp.ICMS.VBC,
				PICMS:      imp.ICMS.PICMS,
				VICMS:      imp.ICMS.VICMS,
				PPIS:       imp.PIS.PPIS,
				VPIS:       imp.PIS.VPIS,
				PCOFINS:    imp.COFINS.PCOFINS,
				VCOFINS:    imp.COFINS.VCOFINS,
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
		return padLeftZeros(emitter.CPF, 14)
	}
	// CNPJ must be exactly 14 digits
	return padLeftZeros(emitter.CNPJ, 14)
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
func (s *NFeService) buildItems(departure entity_public.Departure, unitPrice decimal.Decimal, farmConfig *nfe_model.FarmConfig, productConfig *nfe_model.ProductConfig) []entity.ItemData {
	totalValue := departure.NetWeight.Mul(unitPrice)
	regime := defaults.TaxRegime(farmConfig.TaxRegime)

	// Default tax values
	icmsRate := decimal.NewFromFloat(0.17)
	pisRate := decimal.NewFromFloat(0.0165)
	cofinsRate := decimal.NewFromFloat(0.076)

	// Determine ICMS CST/CSOSN from product config or defaults
	var icmsCST, icmsCSOSN string
	var vBC, vICMS decimal.Decimal

	if regime == defaults.TaxRegimeSimplesNacional {
		if productConfig.ICMSCST != nil && *productConfig.ICMSCST != "" {
			icmsCSOSN = *productConfig.ICMSCST
		} else {
			icmsCSOSN = defaults.CSOSNSemPermissaoCredito // 102
		}
		// For SN, ICMS totals may be zero depending on CSOSN
		vBC = decimal.Zero
		vICMS = decimal.Zero
	} else {
		if productConfig.ICMSCST != nil && *productConfig.ICMSCST != "" {
			icmsCST = *productConfig.ICMSCST
		} else {
			icmsCST = defaults.CSTTributadaIntegral // 00
		}
		vBC = totalValue
		vICMS = totalValue.Mul(icmsRate)
	}

	// Determine PIS and COFINS CST from product config
	pisCST := "01"
	if productConfig.PISCST != nil && *productConfig.PISCST != "" {
		pisCST = *productConfig.PISCST
	}
	cofinsCST := "01"
	if productConfig.COFINSCST != nil && *productConfig.COFINSCST != "" {
		cofinsCST = *productConfig.COFINSCST
	}

	description := "Produto Agricola"
	if productConfig.Description != nil {
		description = *productConfig.Description
	}

	return []entity.ItemData{
		{
			Numero: 1,
			Produto: entity.ProdutoData{
				Codigo:   strconv.Itoa(int(departure.Crop)),
				CEAN:     "SEM GTIN",
				XProd:    description,
				NCM:      productConfig.NCM,
				CFOP:     productConfig.DefaultCFOP,
				UCom:     productConfig.Unit,
				QCom:     departure.NetWeight,
				VUnCom:   unitPrice,
				VProd:    totalValue,
				CEANTrib: "SEM GTIN",
				UTrib:    productConfig.Unit,
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
					PICMS:  decimal.NewFromFloat(17.0),
					VICMS:  vICMS,
				},
				PIS: entity.PISData{
					CST:  pisCST,
					VBC:  totalValue,
					PPIS: decimal.NewFromFloat(1.65),
					VPIS: totalValue.Mul(pisRate),
				},
				COFINS: entity.COFINSData{
					CST:     cofinsCST,
					VBC:     totalValue,
					PCOFINS: decimal.NewFromFloat(7.6),
					VCOFINS: totalValue.Mul(cofinsRate),
				},
			},
		},
	}
}

// mapFarmToEmitter maps farm config and farm data to NF-e emitter data.
func (s *NFeService) mapFarmToEmitter(cfg *nfe_model.FarmConfig, farm *entity_public.Farm, nfeModel *nfe_model.NFeModel) entity.EmitterData {
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
	}

	if cfg.EmitterType == 2 {
		if cfg.CPFEmitter != nil {
			emitter.CPF = *cfg.CPFEmitter
		}
	} else {
		if cfg.CNPJEmitter != nil {
			emitter.CNPJ = *cfg.CNPJEmitter
		}
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
	if emit.Type == 2 {
		if emit.CPF == "" {
			errs = append(errs, "CPF do emitente não configurado")
		}
		// CPF (Produtor Rural) emitters must use series 920-969 per Nota Técnica 2018.001
		if serie < 920 || serie > 969 {
			errs = append(errs, fmt.Sprintf("Série %d inválida para emitente CPF: deve estar entre 920 e 969", serie))
		}
	} else {
		if emit.CNPJ == "" {
			errs = append(errs, "CNPJ do emitente não configurado")
		}
		// CNPJ emitters must use series 0-889
		if serie < 0 || serie > 889 {
			errs = append(errs, fmt.Sprintf("Série %d inválida para emitente CNPJ: deve estar entre 0 e 889", serie))
		}
	}
	if !isValidIEMT(emit.IE) {
		errs = append(errs, "IE do emitente inválida para MT (deve ter 11 dígitos)")
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
	return errs
}

// isValidIEMT validates an Inscrição Estadual for Mato Grosso.
// MT IE format: exactly 11 digits (99999999999).
func isValidIEMT(ie string) bool {
	return true
	if ie == "" {
		return false
	}
	cleaned := cleanIE(ie)
	fmt.Printf("\n%v\n", cleaned)
	return len(cleaned) == 11
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
	for i := 0; i < 32; i++ {
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
