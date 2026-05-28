package entity

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// InvoiceInput holds all data needed to build an NF-e.
type InvoiceInput struct {
	// Identification
	Serie      int
	Numero     int
	NaturezaOp string

	// Emitter
	Emitter EmitterData

	// Recipient
	Recipient RecipientData

	// Product / Items
	Items []ItemData

	// Transport
	Transport TransportData

	// Payment
	Payment PaymentData

	// Totals
	TotalValue decimal.Decimal

	// Additional info
	InformacoesAdicionais string
}

// EmitterData holds emitter (emitente) information.
type EmitterData struct {
	Type       int // 1=CNPJ, 2=CPF
	CNPJ       string
	CPF        string
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
	CFOP     string
	UCom     string
	QCom     decimal.Decimal
	VUnCom   decimal.Decimal
	VProd    decimal.Decimal
	CEANTrib string
	UTrib    string
	QTrib    decimal.Decimal
	VUnTrib  decimal.Decimal
	IndTot   int // 1=item compoe total
}

// ImpostoData holds tax (imposto) information for an item.
type ImpostoData struct {
	ICMS   ICMSData
	PIS    PISData
	COFINS COFINSData
}

// ICMSData holds ICMS tax information.
type ICMSData struct {
	Origem string // 0=Nacional
	CST    string
	ModBC  string
	VBC    decimal.Decimal
	PICMS  decimal.Decimal
	VICMS  decimal.Decimal
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
	VPag   decimal.Decimal
}

// AccessKeyData holds the 44-digit NF-e access key components.
type AccessKeyData struct {
	CUF    string
	AAMM   string
	CNPJ   string
	Mod    string
	Serie  int
	NNF    int
	TpEmis string
	CNF    string
}

// GenerateAccessKey generates the 44-digit access key (chave de acesso).
func GenerateAccessKey(data AccessKeyData) string {
	serieStr := fmt.Sprintf("%03d", data.Serie)
	nnfStr := fmt.Sprintf("%09d", data.NNF)
	key := data.CUF + data.AAMM + data.CNPJ + data.Mod + serieStr + nnfStr + data.TpEmis + data.CNF
	// Add check digit (DV)
	dv := calculateDV(key)
	return key + fmt.Sprintf("%01d", dv)
}

func calculateDV(key string) int {
	weights := []int{4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0
	weightIdx := 0
	for i := len(key) - 1; i >= 0; i-- {
		digit := int(key[i] - '0')
		sum += digit * weights[weightIdx]
		weightIdx++
		if weightIdx >= len(weights) {
			weightIdx = 0
		}
	}
	remainder := sum % 11
	if remainder < 2 {
		return 0
	}
	return 11 - remainder
}
