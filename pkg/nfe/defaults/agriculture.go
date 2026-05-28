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
	NCMMilho = "1005.90.00"
	NCMSoja  = "1201.00.10"
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
			NCM:         "0000.00.00",
			CFOP:        CFOPVendaCompra,
			Unit:        "KG",
			Description: productName,
			ICMSCST:     defaultICMSCST(regime),
			PISCST:      defaultPISCST(regime),
			COFINSCST:   defaultCOFINSCST(regime),
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

// ModeloNFe is the NF-e model code.
const ModeloNFe = "55"

// VersaoLayout is the current NF-e layout version.
const VersaoLayout = "4.00"
