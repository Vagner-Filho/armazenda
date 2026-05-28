package nfe_service

import (
	"fmt"
	"os"
	"strconv"
	"time"

	entity_public "armazenda/entity/public"
	"armazenda/model/departure_model"
	model_error "armazenda/model/error"
	"armazenda/model/nfe_model"
	"armazenda/model/person_model"
	"armazenda/pkg/nfe/config"
	"armazenda/pkg/nfe/defaults"
	"armazenda/pkg/nfe/entity"
	"armazenda/pkg/nfe/service"

	"github.com/shopspring/decimal"
)

// NFeService bridges Armazenda entities with the pkg/nfe package.
type NFeService struct{}

// NewNFeService creates a new NFe service.
func NewNFeService() *NFeService {
	return &NFeService{}
}

// BuildInvoiceFromDeparture builds an NF-e for a departure.
func (s *NFeService) BuildInvoiceFromDeparture(departureID uint32, unitPrice decimal.Decimal, farmID uint32) (string, entity_public.Toast) {
	// Get departure
	dModel := departure_model.GetDepartureModel()
	departure, err := dModel.GetDeparture(departureID)
	if err != nil {
		if err.IsServerErr {
			return "", entity_public.GetErrorToast("Internal error fetching departure", "")
		}
		return "", entity_public.GetWarningToast(err.Message, "")
	}

	// Get farm config
	nfeModel := nfe_model.GetNFeModel()
	farmConfig, dbErr := nfeModel.GetFarmConfig(farmID)
	if dbErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("GetFarmConfig error: %v", dbErr.Error()))
		return "", entity_public.GetErrorToast("Failed to get NFe configuration", "")
	}
	if farmConfig == nil {
		return "", entity_public.GetWarningToast("NFe not configured for this farm", "configure in settings")
	}

	// Get recipient
	var recipient entity.RecipientData
	if departure.Recipient != nil {
		pModel := person_model.GetPersonModel()
		person, personErr := pModel.GetPersonById(*departure.Recipient)
		if personErr != nil {
			model_error.GetLoggerModel().Log(fmt.Sprintf("GetPerson error: %v", personErr.Error()))
			return "", entity_public.GetErrorToast("Failed to get recipient info", "")
		}
		recipient = s.mapPersonToRecipient(person)
	}

	// Get emitter
	emitter := s.mapFarmToEmitter(farmConfig)

	// Allocate invoice number
	number, allocErr := nfeModel.AllocateNumber(farmID, farmConfig.Serie)
	if allocErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("AllocateNumber error: %v", allocErr.Error()))
		return "", entity_public.GetErrorToast("Failed to allocate invoice number", "")
	}

	// Build items
	items := []entity.ItemData{
		{
			Numero: 1,
			Produto: entity.ProdutoData{
				Codigo:   strconv.Itoa(int(departure.Crop)),
				CEAN:     "SEM GTIN",
				XProd:    "Produto Agricola",
				NCM:      defaults.NCMSoja,
				CFOP:     defaults.CFOPVendaProducao,
				UCom:     "KG",
				QCom:     departure.NetWeight,
				VUnCom:   unitPrice,
				VProd:    departure.NetWeight.Mul(unitPrice),
				CEANTrib: "SEM GTIN",
				UTrib:    "KG",
				QTrib:    departure.NetWeight,
				VUnTrib:  unitPrice,
				IndTot:   1,
			},
			Imposto: entity.ImpostoData{
				ICMS: entity.ICMSData{
					Origem: defaults.ICMSOrigemNacional,
					CST:    defaults.CSTTributadaIntegral,
					ModBC:  "3",
					VBC:    departure.NetWeight.Mul(unitPrice),
					PICMS:  decimal.NewFromFloat(17.0),
					VICMS:  departure.NetWeight.Mul(unitPrice).Mul(decimal.NewFromFloat(0.17)),
				},
				PIS: entity.PISData{
					CST:  "01",
					VBC:  departure.NetWeight.Mul(unitPrice),
					PPIS: decimal.NewFromFloat(1.65),
					VPIS: departure.NetWeight.Mul(unitPrice).Mul(decimal.NewFromFloat(0.0165)),
				},
				COFINS: entity.COFINSData{
					CST:     "01",
					VBC:     departure.NetWeight.Mul(unitPrice),
					PCOFINS: decimal.NewFromFloat(7.6),
					VCOFINS: departure.NetWeight.Mul(unitPrice).Mul(decimal.NewFromFloat(0.076)),
				},
			},
		},
	}

	// Build input
	input := entity.InvoiceInput{
		Serie:      farmConfig.Serie,
		Numero:     number,
		NaturezaOp: "Venda de producao do estabelecimento",
		Emitter:    emitter,
		Recipient:  recipient,
		Items:      items,
		Transport: entity.TransportData{
			ModFrete: farmConfig.DefaultModFrete,
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

	// Build and sign
	certPassword := decryptPassword(farmConfig.CertificatePasswordEncrypted)
	sefazCfg := config.SefazConfig{
		Environment: config.Environment(farmConfig.Environment),
		StateUF:     farmConfig.EmitterUF,
		Timeout:     30 * time.Second,
	}
	invService := service.NewInvoiceService(nil, sefazCfg)
	signedXML, signErr := invService.BuildAndSign(input, farmConfig.CertificatePath, certPassword)
	if signErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("BuildAndSign error: %v", signErr.Error()))
		return "", entity_public.GetErrorToast("Failed to build and sign NF-e", signErr.Error())
	}

	// Save to database
	accessKey := entity.GenerateAccessKey(entity.AccessKeyData{
		CUF:    defaults.UFCode(emitter.UF),
		AAMM:   time.Now().Format("0601"),
		CNPJ:   emitter.CNPJ,
		Mod:    defaults.ModeloNFe,
		Serie:  input.Serie,
		NNF:    input.Numero,
		TpEmis: "1",
		CNF:    fmt.Sprintf("%08d", input.Numero),
	})

	_, createErr := nfeModel.CreateInvoice(departureID, accessKey, farmConfig.Serie, number,
		defaults.CFOPVendaProducao, defaults.NCMSoja, departure.NetWeight, unitPrice, input.TotalValue)
	if createErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("CreateInvoice error: %v", createErr.Error()))
		return "", entity_public.GetErrorToast("Failed to save invoice", "")
	}

	return signedXML, entity_public.GetSuccessToast("NF-e generated and signed", "")
}

// mapFarmToEmitter maps farm config to NF-e emitter data.
func (s *NFeService) mapFarmToEmitter(cfg *nfe_model.FarmConfig) entity.EmitterData {
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

	return emitter
}

// mapPersonToRecipient maps a person to NF-e recipient data.
func (s *NFeService) mapPersonToRecipient(person entity_public.PersonDisplay) entity.RecipientData {
	// Simplified mapping - in production, fetch full person data
	return entity.RecipientData{
		Type:       int(person.Type),
		XNome:      person.Name,
		Logradouro: "",
		Numero:     "",
		Bairro:     "",
		CodigoMun:  "",
		Municipio:  "",
		UF:         "",
		CEP:        "",
		IndIEDest:  "9",
	}
}

// decryptPassword decrypts the certificate password using the environment key.
func decryptPassword(encrypted string) string {
	key := os.Getenv("NFE_CERT_KEY")
	if key == "" {
		return encrypted // In development, may be plain text
	}
	// TODO: Implement actual decryption (e.g., AES)
	return encrypted
}
