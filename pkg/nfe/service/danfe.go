package service

import (
	"bytes"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/signintech/gopdf"
)

// DANFEGenerator generates DANFE PDFs.
type DANFEGenerator struct{}

// NewDANFEGenerator creates a new DANFE generator.
func NewDANFEGenerator() *DANFEGenerator {
	return &DANFEGenerator{}
}

// DANFEData holds the data needed to generate a DANFE.
type DANFEData struct {
	AccessKey      string
	EmitterName    string
	EmitterCNPJ    string
	EmitterAddress string
	EmitterCity    string
	EmitterUF      string
	DestName       string
	DestCNPJ       string
	DestAddress    string
	DestCity       string
	DestUF         string
	NaturezaOp     string
	Numero         int
	Serie          int
	EmissionDate   string
	Products       []DANFEProduct
	TotalValue     decimal.Decimal
	ICMSValue      decimal.Decimal
	Protocol       string
	ProtocolDate   string
}

// DANFEProduct holds a single product line for the DANFE.
type DANFEProduct struct {
	Code      string
	Desc      string
	NCM       string
	CFOP      string
	Unit      string
	Quantity  decimal.Decimal
	UnitPrice decimal.Decimal
	Total     decimal.Decimal
}

// Generate creates a DANFE PDF and returns it as bytes.
func (g *DANFEGenerator) Generate(data DANFEData) ([]byte, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	// Use a standard font
	if err := pdf.AddTTFFont("arial", "assets/fonts/arial.ttf"); err != nil {
		// Fallback: use built-in font
		pdf.SetFont("Helvetica", "", 10)
	} else {
		pdf.SetFont("arial", "", 10)
	}

	// Header
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "DANFE - Documento Auxiliar da Nota Fiscal Eletronica")

	pdf.SetXY(10, 20)
	pdf.Cell(nil, fmt.Sprintf("Chave de Acesso: %s", data.AccessKey))

	pdf.SetXY(10, 30)
	pdf.Cell(nil, fmt.Sprintf("Numero: %d  Serie: %d", data.Numero, data.Serie))

	pdf.SetXY(10, 40)
	pdf.Cell(nil, fmt.Sprintf("Natureza da Operacao: %s", data.NaturezaOp))

	// Emitter section
	pdf.SetXY(10, 55)
	pdf.Cell(nil, "EMITENTE")
	pdf.SetXY(10, 65)
	pdf.Cell(nil, fmt.Sprintf("Nome: %s", data.EmitterName))
	pdf.SetXY(10, 75)
	pdf.Cell(nil, fmt.Sprintf("CNPJ: %s", data.EmitterCNPJ))
	pdf.SetXY(10, 85)
	pdf.Cell(nil, fmt.Sprintf("Endereco: %s, %s - %s", data.EmitterAddress, data.EmitterCity, data.EmitterUF))

	// Recipient section
	pdf.SetXY(10, 100)
	pdf.Cell(nil, "DESTINATARIO")
	pdf.SetXY(10, 110)
	pdf.Cell(nil, fmt.Sprintf("Nome: %s", data.DestName))
	pdf.SetXY(10, 120)
	pdf.Cell(nil, fmt.Sprintf("CNPJ/CPF: %s", data.DestCNPJ))
	pdf.SetXY(10, 130)
	pdf.Cell(nil, fmt.Sprintf("Endereco: %s, %s - %s", data.DestAddress, data.DestCity, data.DestUF))

	// Products table header
	y := 150
	pdf.SetXY(10, float64(y))
	pdf.Cell(nil, "Codigo")
	pdf.SetXY(50, float64(y))
	pdf.Cell(nil, "Descricao")
	pdf.SetXY(200, float64(y))
	pdf.Cell(nil, "NCM")
	pdf.SetXY(250, float64(y))
	pdf.Cell(nil, "CFOP")
	pdf.SetXY(290, float64(y))
	pdf.Cell(nil, "Un")
	pdf.SetXY(320, float64(y))
	pdf.Cell(nil, "Qtd")
	pdf.SetXY(360, float64(y))
	pdf.Cell(nil, "V.Unit")
	pdf.SetXY(420, float64(y))
	pdf.Cell(nil, "V.Total")

	// Products
	y += 10
	for _, p := range data.Products {
		pdf.SetXY(10, float64(y))
		pdf.Cell(nil, p.Code)
		pdf.SetXY(50, float64(y))
		pdf.Cell(nil, p.Desc)
		pdf.SetXY(200, float64(y))
		pdf.Cell(nil, p.NCM)
		pdf.SetXY(250, float64(y))
		pdf.Cell(nil, p.CFOP)
		pdf.SetXY(290, float64(y))
		pdf.Cell(nil, p.Unit)
		pdf.SetXY(320, float64(y))
		pdf.Cell(nil, p.Quantity.StringFixed(3))
		pdf.SetXY(360, float64(y))
		pdf.Cell(nil, p.UnitPrice.StringFixed(4))
		pdf.SetXY(420, float64(y))
		pdf.Cell(nil, p.Total.StringFixed(2))
		y += 10
	}

	// Totals
	y += 10
	pdf.SetXY(320, float64(y))
	pdf.Cell(nil, fmt.Sprintf("Total: %s", data.TotalValue.StringFixed(2)))
	y += 10
	pdf.SetXY(320, float64(y))
	pdf.Cell(nil, fmt.Sprintf("ICMS: %s", data.ICMSValue.StringFixed(2)))

	// Protocol
	if data.Protocol != "" {
		y += 20
		pdf.SetXY(10, float64(y))
		pdf.Cell(nil, fmt.Sprintf("Protocolo de Autorizacao: %s", data.Protocol))
		pdf.SetXY(10, float64(y+10))
		pdf.Cell(nil, fmt.Sprintf("Data de Autorizacao: %s", data.ProtocolDate))
	}

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write PDF: %w", err)
	}

	return buf.Bytes(), nil
}
