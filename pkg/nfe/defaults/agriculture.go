package defaults

// TaxRegime represents the Brazilian tax regime.
type TaxRegime int

const (
	TaxRegimeSimplesNacional TaxRegime = 1
	TaxRegimeLucroReal       TaxRegime = 2
	TaxRegimeLucroPresumido  TaxRegime = 3
)

func (t TaxRegime) String() string {
	switch t {
	case TaxRegimeSimplesNacional:
		return "Simples Nacional"
	case TaxRegimeLucroReal:
		return "Lucro Real"
	case TaxRegimeLucroPresumido:
		return "Lucro Presumido"
	default:
		return "Simples Nacional"
	}
}

// CRT (Código de Regime Tributário) returns the NF-e CRT code.
func (t TaxRegime) CRT() string {
	switch t {
	case TaxRegimeSimplesNacional:
		return "1"
	case TaxRegimeLucroReal:
		return "3"
	case TaxRegimeLucroPresumido:
		return "3"
	default:
		return "1"
	}
}

// EmitterType represents the type of NF-e emitter.
type EmitterType int

const (
	EmitterTypeCNPJ EmitterType = 1
	EmitterTypeCPF  EmitterType = 2
)

// NCM codes for agricultural products.
const (
	NCMMilho = "10059010"
	NCMSoja  = "12019000"
)

// CFOP codes for agricultural operations.
const (
	CFOPVendaProducao = "5101" // Venda de producao do estabelecimento
	CFOPVendaCompra   = "5102" // Venda de mercadoria adquirida ou recebida de terceiros
	CFOPRemessa       = "5901" // Remessa para industrializacao
	CFOPDevolucao     = "6202" // Devolucao de compra
)

// CST/CSOSN codes.
const (
	CSTTributadaIntegral      = "00"
	CSTIsenta                 = "40"
	CSTNaoTributada           = "41"
	CSOSSN101Tributada        = "101"
	CSOSSN102SemTributacao    = "102"
	CSOSSN500ICMSSubstituicao = "500"
	CSOSNSemPermissaoCredito  = "102"
)

// ICMS origin codes.
const (
	ICMSOrigemNacional              = "0"
	ICMSOrigemEstrangeiraImportacao = "1"
	ICMSOrigemEstrangeiraAdquirida  = "2"
	ICMSOrigemNacionalImportada     = "3"
	ICMSOrigemNacionalImportada40   = "4"
	ICMSOrigemNacionalImportada50   = "5"
	ICMSOrigemEstrangeiraImportada  = "6"
	ICMSOrigemNacionalImportada60   = "7"
)

// ModFrete (modalidade do frete).
type ModFrete int

const (
	ModFretePorContaEmitente     ModFrete = 0 // CIF
	ModFretePorContaDestinatario ModFrete = 1 // FOB
	ModFretePorContaTerceiros    ModFrete = 2
	ModFreteProprioRemetente     ModFrete = 3
	ModFreteProprioDestinatario  ModFrete = 4
	ModFreteSemFrete             ModFrete = 9
)

func (m ModFrete) String() string {
	switch m {
	case ModFretePorContaEmitente:
		return "0"
	case ModFretePorContaDestinatario:
		return "1"
	case ModFretePorContaTerceiros:
		return "2"
	case ModFreteProprioRemetente:
		return "3"
	case ModFreteProprioDestinatario:
		return "4"
	case ModFreteSemFrete:
		return "9"
	default:
		return "1"
	}
}

// ProductDefaults holds the default fiscal configuration for a product.
type ProductDefaults struct {
	NCM         string
	CFOP        string
	Unit        string
	Description string
	ICMSCST     string
	PISCST      string
	COFINSCST   string
	// Tax reform (IBS/CBS) defaults applied when the farm config does not
	// override them. CClassTrib is the NT 2025.002-RTC classification code
	// that ties IBS/CBS to a specific tax regime / sector rule.
	IBSCBSCST   string
	CClassTrib  string
}

// GetProductDefaults returns the default fiscal configuration for a product.
func GetProductDefaults(productName string, regime TaxRegime) ProductDefaults {
	switch productName {
	case "Milho":
		return productDefaultsMilho(regime)
	case "Soja":
		return productDefaultsSoja(regime)
	default:
		return ProductDefaults{
			NCM:         "00000000",
			CFOP:        CFOPVendaCompra,
			Unit:        "KG",
			Description: productName,
			ICMSCST:     defaultICMSCST(regime),
			PISCST:      defaultPISCST(regime),
			COFINSCST:   defaultCOFINSCST(regime),
			IBSCBSCST:   IBSCBSCSTTributadaIntegral,
			CClassTrib:  CClassTribDefault,
		}
	}
}

func productDefaultsMilho(regime TaxRegime) ProductDefaults {
	return ProductDefaults{
		NCM:         NCMMilho,
		CFOP:        CFOPVendaProducao,
		Unit:        "KG",
		Description: "Milho em grao",
		ICMSCST:     defaultICMSCST(regime),
		PISCST:      defaultPISCST(regime),
		COFINSCST:   defaultCOFINSCST(regime),
		IBSCBSCST:   IBSCBSCSTTributadaIntegral,
		CClassTrib:  CClassTribDefault,
	}
}

func productDefaultsSoja(regime TaxRegime) ProductDefaults {
	return ProductDefaults{
		NCM:         NCMSoja,
		CFOP:        CFOPVendaProducao,
		Unit:        "KG",
		Description: "Soja em grao",
		ICMSCST:     defaultICMSCST(regime),
		PISCST:      defaultPISCST(regime),
		COFINSCST:   defaultCOFINSCST(regime),
		IBSCBSCST:   IBSCBSCSTTributadaIntegral,
		CClassTrib:  CClassTribDefault,
	}
}

func defaultICMSCST(regime TaxRegime) string {
	if regime == TaxRegimeSimplesNacional {
		return CSOSSN101Tributada
	}
	return CSTTributadaIntegral
}

func defaultPISCST(regime TaxRegime) string {
	return "01" // Operacao tributavel com aliquota basica
}

func defaultCOFINSCST(regime TaxRegime) string {
	return "01" // Operacao tributavel com aliquota basica
}

// UFCode returns the IBGE state code.
func UFCode(uf string) string {
	codes := map[string]string{
		"RO": "11", "AC": "12", "AM": "13", "RR": "14", "PA": "15", "AP": "16", "TO": "17",
		"MA": "21", "PI": "22", "CE": "23", "RN": "24", "PB": "25", "PE": "26", "AL": "27", "SE": "28", "BA": "29",
		"MG": "31", "ES": "32", "RJ": "33", "SP": "35",
		"PR": "41", "SC": "42", "RS": "43",
		"MS": "50", "MT": "51", "GO": "52", "DF": "53",
	}
	if code, ok := codes[uf]; ok {
		return code
	}
	return ""
}

// TpEmis represents the NF-e emission type (tpEmis).
type TpEmis int

const (
	EmissaoNormal TpEmis = 1
	EPEC          TpEmis = 4 // reserved for future use
	FSDA          TpEmis = 5 // reserved (never active for SaaS)
	SVCAN         TpEmis = 6
	SVCRS         TpEmis = 7
)

func (t TpEmis) String() string {
	switch t {
	case EmissaoNormal:
		return "1"
	case EPEC:
		return "4"
	case FSDA:
		return "5"
	case SVCAN:
		return "6"
	case SVCRS:
		return "7"
	default:
		return "1"
	}
}

// IsContingency returns true for any contingency emission mode.
func (t TpEmis) IsContingency() bool {
	return t == EPEC || t == FSDA || t == SVCAN || t == SVCRS
}

// SVCForState returns the SVC emission type for a given state per Ato COTEPE 39/2012.
// Returns EmissaoNormal for states not mapped to any SVC.
func SVCForState(uf string) TpEmis {
	switch uf {
	case "AC", "AL", "AP", "MG", "PB", "RJ", "RS", "RO", "RR", "SC", "SE", "SP", "TO", "DF":
		return SVCAN
	case "AM", "BA", "CE", "ES", "GO", "MA", "MT", "MS", "PA", "PE", "PI", "PR", "RN":
		return SVCRS
	default:
		return EmissaoNormal
	}
}

// ModeloNFe is the NF-e model code.
const ModeloNFe = "55"

// VersaoLayout is the current NF-e layout version.
const VersaoLayout = "4.00"

// NaturezaOpForCFOP returns the standard natureza da operação description
// for a given CFOP code. If the CFOP is not in the mapping, returns a
// generic "Venda de mercadoria".
func NaturezaOpForCFOP(cfop string) string {
	switch cfop {
	case "1101", "5101", "7101":
		return "Venda de producao do estabelecimento"
	case "1102", "5102", "7102":
		return "Venda de mercadoria adquirida ou recebida de terceiros"
	case "1901", "5901", "7901":
		return "Remessa para industrializacao"
	case "1202", "2202", "5202", "6202":
		return "Devolucao de compra"
	case "1103", "5103", "7103":
		return "Venda de producao do estabelecimento ao contribuinte do ICMS"
	case "1104", "5104", "7104":
		return "Venda de producao do estabelecimento a nao contribuinte"
	default:
		return "Venda de mercadoria"
	}
}
