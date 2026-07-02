package entity

import "github.com/shopspring/decimal"

// DANFEData holds the data needed to generate a SEFAZ-compliant DANFE.
type DANFEData struct {
	// Identification
	AccessKey    string
	NaturezaOp   string
	Numero       int
	Serie        int
	EmissionDate string

	// Emission context (Grupo B)
	TpEmis   string // B22: 1=normal, 4=EPEC, 5=FS-DA, 6=SVC-AN, 7=SVC-RS
	TpAmb    string // B24: 1=produção, 2=homologação
	TpNF     string // B11: 0=Entrada, 1=Saída
	DhSaiEnt string // B10: data/hora de saída ou entrada
	DhCont   string // B28: data/hora da entrada em contingência (formatted)
	XJust    string // B29: justificativa da entrada em contingência
	VerProc  string // B27: versão do processo de emissão

	// Emitter
	EmitterName         string
	EmitterCNPJ         string
	EmitterIE           string
	EmitterCRT          string
	EmitterAddress      string
	EmitterNumber       string
	EmitterComplement   string
	EmitterNeighborhood string
	EmitterCEP          string
	EmitterCity         string
	EmitterUF           string
	EmitterPhone        string

	// Destinatario
	DestName         string
	DestCNPJ         string
	DestIE           string
	DestIndIEDest    string
	DestAddress      string
	DestNumber       string
	DestComplement   string
	DestNeighborhood string
	DestCEP          string
	DestCity         string
	DestUF           string
	DestPhone        string

	// Products
	Products []DANFEProduct

	// Totals (ICMS)
	TotalValue decimal.Decimal
	VBC        decimal.Decimal
	VICMS      decimal.Decimal
	VICMSDeson decimal.Decimal
	VBCST      decimal.Decimal
	VST        decimal.Decimal
	VII        decimal.Decimal
	VIPI       decimal.Decimal
	VPIS       decimal.Decimal
	VCOFINS    decimal.Decimal
	VFrete     decimal.Decimal
	VSeg       decimal.Decimal
	VDesc      decimal.Decimal
	VOutro     decimal.Decimal
	VTotTrib   decimal.Decimal

	// Transport
	ModFrete      string
	TranspName    string
	TranspCNPJ    string
	TranspIE      string
	TranspAddress string
	TranspCity    string
	TranspUF      string
	QVol          string
	Esp           string
	Marca         string
	NVol          string
	PesoL         decimal.Decimal
	PesoB         decimal.Decimal
	VeicPlate     string
	VeicUF        string

	// ISSQN (conditional)
	VBCISSQN      decimal.Decimal
	VISSQN        decimal.Decimal
	VPISISSQN     decimal.Decimal
	VCOFINSSISSQN decimal.Decimal

	// Additional info
	InfCpl     string
	InfAdFisco string

	// Protocol (authorized only)
	Protocol     string
	ProtocolDate string
	CStat        string
	XMotivo      string
}

// DANFEProduct holds a single product line for the DANFE.
type DANFEProduct struct {
	Code      string
	Desc      string
	NCM       string
	CST       string
	CFOP      string
	Unit      string
	Quantity  decimal.Decimal
	UnitPrice decimal.Decimal
	Total     decimal.Decimal
	// Tributable unit (Grupo I): required to be shown when different from
	// the commercial unit (MOC 7.0 Anexo II §3.1.7).
	UTrib     string
	QTrib     decimal.Decimal
	VUnTrib   decimal.Decimal
	InfAdProd string
	// Taxes
	VBC     decimal.Decimal
	PICMS   decimal.Decimal
	VICMS   decimal.Decimal
	PIPI    decimal.Decimal
	VIPI    decimal.Decimal
	PPIS    decimal.Decimal
	VPIS    decimal.Decimal
	PCOFINS decimal.Decimal
	VCOFINS decimal.Decimal
	VFrete  decimal.Decimal
	VSeg    decimal.Decimal
	VDesc   decimal.Decimal
	VOutro  decimal.Decimal
}
