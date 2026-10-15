package entity

import (
	"fmt"
	"time"

	"armazenda/pkg/nfe/defaults"

	"github.com/shopspring/decimal"
)

// InvoiceInput holds all data needed to build an NF-e.
type InvoiceInput struct {
	// Identification
	Serie      int
	Numero     int
	CNF        string // Random 8-digit numeric code for access key
	NaturezaOp string

	// Environment: 1=production, 2=homologation
	Environment int

	// Emission type: 1=normal, 6=SVC-AN, 7=SVC-RS
	TpEmis defaults.TpEmis

	// Contingency fields (required when TpEmis != 1)
	DhCont *time.Time
	XJust  string

	// Exit/entry date-time (dhSaiEnt, B15a). Nil omits the element; the
	// caller owns the DhSaiEnt >= DhEmi validation.
	DhSaiEnt *time.Time

	// Finalidade da emissão (finNFe): "1"=normal, "2"=complementar,
	// "3"=ajuste, "4"=devolução. Empty defaults to "1" in the builder.
	FinNFe string

	// Referenced NF-e access keys (NFref/refNFe) — devoluções must
	// reference the originating NF-e. Invalid entries (not 44 digits) are
	// skipped by the builder.
	NFRefs []string

	// Emitter
	Emitter EmitterData

	// Recipient
	Recipient RecipientData

	// Delivery-site group (vendas a ordem). Nil omits <entrega>.
	Entrega *EntregaData

	// Technical responsible (<infRespTec>, NT 2022.002 / E28a). Nil omits
	// the block, but SEFAZ rejects production emissions without it.
	RespTec *RespTecData

	// Product / Items
	Items []ItemData

	// Transport
	Transport TransportData

	// Payment
	Payment PaymentData
	// Billing/fatura. Nil preserves the builder's legacy behavior (emit
	// <cobr>/<fat> for non-zero totals); an explicit value can suppress it.
	Cobranca *CobrancaData

	// Totals
	TotalValue decimal.Decimal

	// Additional info
	InformacoesAdicionais string
}

// EmitterData holds emitter (emitente) information.
type EmitterData struct {
	Type       int    // 1=CNPJ, 2=CPF
	Document   string // CNPJ or CPF
	IE         string
	XNome      string
	XFant      string
	Logradouro string
	Numero     string
	Bairro     string
	CodigoMun  string
	Municipio  string
	UF         string
	CEP        string
	Fone       string
	CRT        string // 1=Simples Nacional, 3=Normal
}

// RecipientData holds recipient (destinatario) information.
type RecipientData struct {
	Type       int // 1=CNPJ, 2=CPF
	CNPJ       string
	CPF        string
	IE         string
	Email      string
	XNome      string
	Logradouro string
	Numero     string
	Bairro     string
	CodigoMun  string
	Municipio  string
	UF         string
	CEP        string
	Fone       string
	IndIEDest  string // 1=Contribuinte ICMS, 2=Isento, 9=Nao contribuinte
}

// ItemData holds a single item (det) in the NF-e.
type ItemData struct {
	Numero    int
	Produto   ProdutoData
	Imposto   ImpostoData
	InfAdProd string
}

// ProdutoData holds product (prod) information.
type ProdutoData struct {
	Codigo   string
	CEAN     string
	XProd    string
	NCM      string
	CEST     string // Código Especificador da Substituição Tributária
	CFOP     string
	UCom     string
	QCom     decimal.Decimal
	VUnCom   decimal.Decimal
	VProd    decimal.Decimal
	CEANTrib string
	UTrib    string
	QTrib    decimal.Decimal
	VUnTrib  decimal.Decimal
	// VDesc is the item-level discount (prod/vDesc, G3 devoluções). Zero
	// omits the element and feeds the ICMSTot/vDesc total.
	VDesc  decimal.Decimal
	IndTot int // 1=item compoe total
}

// ImpostoData holds tax (imposto) information for an item.
type ImpostoData struct {
	ICMS   ICMSData
	PIS    PISData
	COFINS COFINSData
	// IBSCBS carries the IBS/CBS groups required by the 2026 indirect tax
	// reform (NT 2025.002-RTC / MOC 7.0). When IsTaxReformActive is false the
	// XML builder still emits a zero-valued <IBSCBS> so the schema stays stable
	// across years; downstream consumers should treat zero rates as "not yet
	// reform-active".
	IBSCBS IBSCBSData
}

// IBSCBSData holds the per-item IBS + CBS values for the indirect tax reform.
// CST is the 3-digit IBSCBS classification (see defaults.IBSCBSCST* constants).
// CClassTrib is the NT 2025.002-RTC regime classification. VBC is the tax
// base (same as vProd for grain sales). VIBSUF and VIBSMun are the state and
// municipal shares of the per-item IBS; VIBS is the per-item total
// (VIBSUF + VIBSMun) — the SEFAZ schema requires both the split values
// inside <gIBSUF>/<gIBSMun> and the total inside <gIBS> at this nesting
// level. PCBS/VCBS are the federal CBS rate and amount.
//
// Rates are stored as decimal *rates* (0.009 = 0.9 %); the XML builder
// multiplies by 100 to produce pIBSUF / pCBS percentage strings.
type IBSCBSData struct {
	CST        string
	CClassTrib string
	VBC        decimal.Decimal
	VIBSUF     decimal.Decimal
	VIBSMun    decimal.Decimal
	VIBS       decimal.Decimal
	PCBS       decimal.Decimal
	VCBS       decimal.Decimal
}

// TaxRates holds tax rate inputs for an emission.
// A nil pointer means "not provided" (the resolver should fall back to the
// product config or a last-resort default). A non-nil pointer — including a
// pointer to a zero Decimal — is authoritative; a zero value means the user
// explicitly declared the operation exempt for that tax.
type TaxRates struct {
	ICMSRate   *decimal.Decimal
	PISRate    *decimal.Decimal
	COFINSRate *decimal.Decimal
	// IBSRate / CBSRate are the per-item decimal rates (e.g. 0.001 = 0.1 %
	// IBS, 0.009 = 0.9 % CBS). Nil means "fall back to the farm config /
	// 2026 default".
	IBSRate *decimal.Decimal
	CBSRate *decimal.Decimal
}

// InvoiceOverrides holds per-emission user overrides that differ from the
// farm config defaults. Nil pointers mean "use the farm config / product default".
type InvoiceOverrides struct {
	NaturezaOp  *string
	ProductDesc *string
	NCM         *string
	CEST        *string
	Unit        *string
	ModFrete    *int
	ICMSCST     *string
	PISCST      *string
	COFINSCST   *string
	// Tax reform overrides (IBS/CBS). All optional; nil falls back to farm
	// config (which itself defaults to the 2026 symbolic rates).
	CBSCST     *string
	IBSCST     *string
	CClassTrib *string
	InfCpl     *string // Informações complementares; overrides the auto-generated CND text
	// Departure-only inputs. Detached NF-e does not consume this field.
	Departure *DepartureInvoiceOverrides
}

// DepartureInvoiceOverrides carries form values specific to romaneio-linked
// NF-e emission. Nil inclusion flags preserve the existing departure-derived
// defaults for old clients; explicit false suppresses that generated group.
type DepartureInvoiceOverrides struct {
	TPag            string
	IndPag          int
	IncludeCobranca *bool
	Transport       DepartureTransportInput
}

// DepartureTransportInput is the optional form payload for Grupo X. The
// service validates it and merges it with the existing departure-derived
// vehicle/volume defaults.
type DepartureTransportInput struct {
	TransporterName     string
	TransporterDocument string
	TransporterIE       string
	TransporterStreet   string
	TransporterNumber   string
	TransporterCity     string
	TransporterUF       string

	VehiclePlate string
	VehicleUF    string
	VehicleRNTC  string

	IncludeVehicle *bool
	IncludeVolumes *bool
	VolumeQuantity *int
	VolumeSpecies  string
	VolumeBrand    string
	VolumeNet      *decimal.Decimal
	VolumeGross    *decimal.Decimal
}

// InvoiceEmissionInputs is the resolved, per-invoice snapshot needed to
// reproduce payment, billing and transport XML during draft/SVC retries.
// Older invoice rows have no snapshot and continue to use legacy defaults.
type InvoiceEmissionInputs struct {
	Payment         PaymentData   `json:"payment"`
	IncludeCobranca bool          `json:"include_cobranca"`
	Transport       TransportData `json:"transport"`
}

// ICMSData holds ICMS tax information.
// For Simples Nacional, use CSOSN instead of CST, and leave VBC/PICMS/VICMS zero.
//
// The extra regime-normal fields below back the CST variants beyond ICMS00
// (real-world corpus shapes — see pkg/nfe/xml/builder.go buildICMS):
//   - PRedBC feeds <pRedBC> inside ICMS20/ICMS51/ICMS90 (base reduction).
//   - ModBCST/VBCST/PICMSST/VICMSST feed the ICMS90 substitution terms.
//
// They are zero-valued in today's Armazenda flows; ICMS40 ignores them.
type ICMSData struct {
	Origem string // 0=Nacional
	CST    string
	CSOSN  string // Simples Nacional CSOSN code
	ModBC  string
	VBC    decimal.Decimal
	PICMS  decimal.Decimal
	VICMS  decimal.Decimal

	PRedBC  decimal.Decimal // pRedBC (40)
	ModBCST string          // modBCST (72)
	VBCST   decimal.Decimal // vBCST (74)
	PICMSST decimal.Decimal // pICMSST (75)
	VICMSST decimal.Decimal // vICMSST (76)
}

// PISData holds PIS tax information.
type PISData struct {
	CST  string
	VBC  decimal.Decimal
	PPIS decimal.Decimal
	VPIS decimal.Decimal
}

// COFINSData holds COFINS tax information.
type COFINSData struct {
	CST     string
	VBC     decimal.Decimal
	PCOFINS decimal.Decimal
	VCOFINS decimal.Decimal
}

// RespTecData holds the technical responsible (infRespTec, NT 2022.002
// / E28a) — mandatory in production emissions.
type RespTecData struct {
	CNPJ     string
	XContato string
	Email    string
	Fone     string
}

// EntregaData holds the delivery-site group (<entrega>), used by vendas a
// ordem flows. Flat group (CNPJ-or-CPF chosen by Type), sibling of <dest>.
type EntregaData struct {
	Type       int // 1=CNPJ, 2=CPF
	CNPJ       string
	CPF        string
	XNome      string
	Logradouro string
	Numero     string
	Bairro     string
	CodigoMun  string
	Municipio  string
	UF         string
	CEP        string
	IE         string
}

// TransportData holds transport (transp) information.
type TransportData struct {
	ModFrete       int
	Transportadora *TransportadoraData
	Veiculo        *VeiculoData
	Volumes        []VolumeData
}

// TransportadoraData holds transporter information.
type TransportadoraData struct {
	Type      int // 1=CNPJ, 2=CPF
	CNPJ      string
	CPF       string
	XNome     string
	IE        string
	Endereco  string
	UF        string
	Municipio string
}

// VeiculoData holds vehicle information.
type VeiculoData struct {
	Placa string
	UF    string
	RNTC  string
}

// VolumeData holds volume information.
type VolumeData struct {
	QVol  int
	Esp   string
	Marca string
	NVol  string
	PesoL decimal.Decimal
	PesoB decimal.Decimal
}

// PaymentData holds payment (pag) information.
type PaymentData struct {
	IndPag   int // 0=Pagamento a vista, 1=Pagamento a prazo, 2=Outros
	Detalhes []PagamentoDetalhe
}

// PagamentoDetalhe holds a single payment detail.
type PagamentoDetalhe struct {
	IndPag int    // 0=a vista
	TPag   string // 90=Sem pagamento (common for grain sales with deferred payment)
	XPag   string // free-text mechanic description (detPag/xPag, optional)
	VPag   decimal.Decimal
}

// CobrancaData controls whether the optional <cobr>/<fat> group is emitted.
// The departure flow derives its number and amounts from the NF-e itself.
type CobrancaData struct {
	Enabled bool
}

// AccessKeyData holds the 44-digit NF-e access key components.
type AccessKeyData struct {
	CUF      string
	AAMM     string
	Document string // CNPJ or zero-padded CPF (always 14 digits)
	Mod      string
	Serie    int
	NNF      int
	TpEmis   string
	CNF      string
}

// GenerateAccessKey generates the 44-digit access key (chave de acesso).
func GenerateAccessKey(data AccessKeyData) string {
	serieStr := fmt.Sprintf("%03d", data.Serie)
	nnfStr := fmt.Sprintf("%09d", data.NNF)
	key := data.CUF + data.AAMM + data.Document + data.Mod + serieStr + nnfStr + data.TpEmis + data.CNF
	// Add check digit (DV)
	dv := calculateDV(key)
	return key + fmt.Sprintf("%01d", dv)
}

func calculateDV(key string) int {
	// SEFAZ DV algorithm: multiply digits from RIGHT to LEFT
	// with weights [2,3,4,5,6,7,8,9] cycling.
	sum := 0
	peso := 2
	for i := len(key) - 1; i >= 0; i-- {
		digit := int(key[i] - '0')
		sum += digit * peso
		peso++
		if peso > 9 {
			peso = 2
		}
	}
	remainder := sum % 11
	if remainder < 2 {
		return 0
	}
	return 11 - remainder
}
